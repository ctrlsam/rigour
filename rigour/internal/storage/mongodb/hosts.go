package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ctrlsam/rigour/internal/storage"
	"github.com/ctrlsam/rigour/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type HostRepository struct {
	collection *mongo.Collection
}

func NewHostsRepository(ctx context.Context, coll *mongo.Collection) (storage.HostRepository, error) {
	_, _ = coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "ip", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("ip_unique"),
		},
		{
			Keys: bson.D{
				{Key: "ip", Value: 1},
				{Key: "services.port", Value: 1},
				{Key: "services.protocol", Value: 1},
				{Key: "services.transport", Value: 1},
			},
			Options: options.Index().SetName("services_lookup"),
		},
	})
	return &HostRepository{collection: coll}, nil
}

func (repo *HostRepository) EnsureHost(ctx context.Context, ip string, now time.Time) error {
	if now.IsZero() {
		now = time.Now()
	}

	filter := bson.M{"ip": ip}
	update := bson.M{
		"$setOnInsert": bson.M{
			"ip":         ip,
			"first_seen": now,
			"services":   []types.Service{},
			"labels":     []string{},
		},
		"$set": bson.M{
			"last_seen": now,
		},
	}

	_, err := repo.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return err
}

func (repo *HostRepository) UpdateHost(ctx context.Context, host types.Host) error {
	now := time.Now()
	set := bson.M{
		"last_seen": now,
	}

	if host.ASN != nil {
		set["asn"] = host.ASN
	}
	if host.Location != nil {
		set["location"] = host.Location
	}
	if host.Labels != nil {
		set["labels"] = host.Labels
	}
	if host.IPInt != 0 {
		set["ip_int"] = host.IPInt
	}

	_, err := repo.collection.UpdateOne(ctx, bson.M{"ip": host.IP}, bson.M{"$set": set})
	return err
}

func (repo *HostRepository) GetByIP(ctx context.Context, ip string) (*types.Host, error) {
	if repo == nil || repo.collection == nil {
		return nil, errors.New("mongodb: hosts repository is nil")
	}

	filter := bson.M{"ip": ip}
	var host types.Host
	err := repo.collection.FindOne(ctx, filter).Decode(&host)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("host not found: %s", ip)
		}
		return nil, fmt.Errorf("mongodb: failed to fetch host: %w", err)
	}

	return &host, nil
}

func (repo *HostRepository) UpsertService(ctx context.Context, svc types.Service) error {
	now := svc.LastScan
	if now.IsZero() {
		now = time.Now()
	}

	// Replace existing service or push if missing
	filter := bson.M{
		"ip": svc.IP,
		"services": bson.M{
			"$elemMatch": bson.M{
				"port": svc.Port,
			},
		},
	}
	updateExisting := bson.M{
		"$set": bson.M{
			"services.$": svc,
			"last_seen":  now,
		},
	}

	res, err := repo.collection.UpdateOne(ctx, filter, updateExisting)
	if err != nil {
		return err
	}
	if res.MatchedCount > 0 {
		return nil
	}

	// Not found, push it.
	filterHost := bson.M{"ip": svc.IP}
	pushUpdate := bson.M{
		"$set": bson.M{
			"last_seen": now,
		},
		"$push": bson.M{"services": svc},
	}
	_, err = repo.collection.UpdateOne(ctx, filterHost, pushUpdate)
	if err != nil {
		return fmt.Errorf("mongodb: push service: %w", err)
	}
	return nil
}

func (repo *HostRepository) Search(ctx context.Context, filter map[string]interface{}, lastID string, limit int) ([]types.Host, string, error) {
	if repo == nil || repo.collection == nil {
		return nil, "", errors.New("mongodb: hosts repository is nil")
	}

	// Build the match stage from the filter
	matchStage := bson.M{}
	if len(filter) > 0 {
		matchStage = bson.M(filter)
	}

	// Build the pipeline
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: matchStage}},
	}

	// If lastID is provided, add a filter to skip past it
	if lastID != "" {
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.M{"_id": bson.M{"$gt": lastID}}},
		})
	}

	// Sort by _id and limit to get one extra to check if there are more results
	pipeline = append(pipeline,
		bson.D{{Key: "$sort", Value: bson.M{"_id": 1}}},
		bson.D{{Key: "$limit", Value: limit + 1}},
	)

	cursor, err := repo.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, "", fmt.Errorf("mongodb: search aggregation failed: %w", err)
	}
	defer cursor.Close(ctx)

	var hosts []types.Host
	if err := cursor.All(ctx, &hosts); err != nil {
		return nil, "", fmt.Errorf("mongodb: failed to decode results: %w", err)
	}

	// Determine if there are more results
	var nextID string
	if len(hosts) > limit {
		// There are more results, trim to limit and set next ID
		hosts = hosts[:limit]
		nextID = hosts[len(hosts)-1].ID
	}

	return hosts, nextID, nil
}

func (repo *HostRepository) Facets(ctx context.Context, filter map[string]interface{}) (*storage.FacetCounts, error) {
	if repo == nil || repo.collection == nil {
		return nil, errors.New("mongodb: hosts repository is nil")
	}

	// Build the match stage from the filter
	matchStage := bson.M{}
	if len(filter) > 0 {
		matchStage = bson.M(filter)
	}

	// Build the aggregation pipeline with facets
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: matchStage}},
		bson.D{
			{Key: "$facet", Value: bson.M{
				"services": bson.A{
					bson.M{"$unwind": "$services"},
					bson.M{"$group": bson.M{
						"_id":   "$services.protocol",
						"count": bson.M{"$sum": 1},
					}},
					bson.M{"$sort": bson.M{"count": -1}},
				},
				"countries": bson.A{
					bson.M{"$match": bson.M{"asn.country": bson.M{"$exists": true, "$ne": nil}}},
					bson.M{"$group": bson.M{
						"_id":   "$asn.country",
						"count": bson.M{"$sum": 1},
					}},
					bson.M{"$sort": bson.M{"count": -1}},
				},
				"asns": bson.A{
					bson.M{"$match": bson.M{"asn.number": bson.M{"$exists": true, "$ne": nil}}},
					bson.M{"$group": bson.M{
						"_id": bson.M{
							"number":       "$asn.number",
							"organization": "$asn.organization",
						},
						"count": bson.M{"$sum": 1},
					}},
					bson.M{"$sort": bson.M{"count": -1}},
				},
			}},
		},
	}

	cursor, err := repo.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("mongodb: facet aggregation failed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("mongodb: failed to decode facet results: %w", err)
	}

	if len(results) == 0 {
		return &storage.FacetCounts{
			Services:  make(map[string]int),
			Countries: make(map[string]int),
			ASNs:      make(map[string]int),
		}, nil
	}

	facetResult := results[0]
	counts := &storage.FacetCounts{
		Services:  make(map[string]int),
		Countries: make(map[string]int),
		ASNs:      make(map[string]int),
	}

	// Process services facet
	if servicesFacet, ok := facetResult["services"]; ok {
		if servicesArray, ok := servicesFacet.(bson.A); ok {
			for _, item := range servicesArray {
				if doc, ok := item.(bson.M); ok {
					if id, ok := doc["_id"]; ok {
						if count, ok := doc["count"].(int32); ok {
							counts.Services[fmt.Sprintf("%v", id)] = int(count)
						}
					}
				}
			}
		}
	}

	// Process countries facet
	if countriesFacet, ok := facetResult["countries"]; ok {
		if countriesArray, ok := countriesFacet.(bson.A); ok {
			for _, item := range countriesArray {
				if doc, ok := item.(bson.M); ok {
					if id, ok := doc["_id"]; ok {
						if count, ok := doc["count"].(int32); ok {
							counts.Countries[fmt.Sprintf("%v", id)] = int(count)
						}
					}
				}
			}
		}
	}

	// Process ASNs facet
	if asnsFacet, ok := facetResult["asns"]; ok {
		if asnsArray, ok := asnsFacet.(bson.A); ok {
			for _, item := range asnsArray {
				if doc, ok := item.(bson.M); ok {
					if id, ok := doc["_id"]; ok {
						if count, ok := doc["count"].(int32); ok {
							// Format ASN key as "number-organization"
							if idMap, ok := id.(bson.M); ok {
								number := idMap["number"]
								organization := idMap["organization"]
								key := fmt.Sprintf("%v-%v", number, organization)
								counts.ASNs[key] = int(count)
							}
						}
					}
				}
			}
		}
	}

	return counts, nil
}

var _ storage.HostRepository = (*HostRepository)(nil)
