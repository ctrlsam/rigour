package mongodb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ctrlsam/rigour/internal/storage"
	"github.com/ctrlsam/rigour/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type HostRepository struct {
	collection   *mongo.Collection
	vulnCollName string
}

func NewHostsRepository(ctx context.Context, coll *mongo.Collection, vulnCollName string) (storage.HostRepository, error) {
	if _, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
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
	}); err != nil {
		slog.Warn("failed to create hosts indexes", "collection", coll.Name(), "err", err)
	}
	return &HostRepository{collection: coll, vulnCollName: vulnCollName}, nil
}

// vulnLookupStages returns aggregation stages that join and summarise vulnerabilities for each host.
func (repo *HostRepository) vulnLookupStages() []bson.D {
	severityReduce := bson.M{
		"$reduce": bson.M{
			"input":        "$_vulns",
			"initialValue": "",
			"in": bson.M{
				"$switch": bson.M{
					"branches": bson.A{
						bson.M{"case": bson.M{"$eq": bson.A{"$$this.severity", "critical"}}, "then": "critical"},
						bson.M{"case": bson.M{"$and": bson.A{
							bson.M{"$eq": bson.A{"$$this.severity", "high"}},
							bson.M{"$ne": bson.A{"$$value", "critical"}},
						}}, "then": "high"},
						bson.M{"case": bson.M{"$and": bson.A{
							bson.M{"$eq": bson.A{"$$this.severity", "medium"}},
							bson.M{"$not": bson.A{bson.M{"$in": bson.A{"$$value", bson.A{"critical", "high"}}}}},
						}}, "then": "medium"},
						bson.M{"case": bson.M{"$and": bson.A{
							bson.M{"$eq": bson.A{"$$this.severity", "low"}},
							bson.M{"$not": bson.A{bson.M{"$in": bson.A{"$$value", bson.A{"critical", "high", "medium"}}}}},
						}}, "then": "low"},
						bson.M{"case": bson.M{"$and": bson.A{
							bson.M{"$eq": bson.A{"$$this.severity", "info"}},
							bson.M{"$eq": bson.A{"$$value", ""}},
						}}, "then": "info"},
					},
					"default": "$$value",
				},
			},
		},
	}

	cveIDsReduce := bson.M{
		"$reduce": bson.M{
			"input":        "$_vulns",
			"initialValue": bson.A{},
			"in": bson.M{
				"$setUnion": bson.A{
					"$$value",
					bson.M{"$ifNull": bson.A{"$$this.cve_ids", bson.A{}}},
				},
			},
		},
	}

	return []bson.D{
		{{Key: "$lookup", Value: bson.M{
			"from":         repo.vulnCollName,
			"localField":   "ip",
			"foreignField": "ip",
			"as":           "_vulns",
		}}},
		{{Key: "$addFields", Value: bson.M{
			"vulns": bson.M{
				"$cond": bson.M{
					"if":   bson.M{"$gt": bson.A{bson.M{"$size": "$_vulns"}, 0}},
					"then": bson.M{"count": bson.M{"$size": "$_vulns"}, "max_severity": severityReduce, "cve_ids": cveIDsReduce},
					"else": "$$REMOVE",
				},
			},
		}}},
		{{Key: "$project", Value: bson.M{"_vulns": 0}}},
	}
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

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"ip": ip}}},
	}
	pipeline = append(pipeline, repo.vulnLookupStages()...)
	pipeline = append(pipeline, bson.D{{Key: "$limit", Value: 1}})

	cursor, err := repo.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("mongodb: failed to fetch host: %w", err)
	}
	defer cursor.Close(ctx)

	hosts := []types.Host{}
	if err := cursor.All(ctx, &hosts); err != nil {
		return nil, fmt.Errorf("mongodb: failed to decode host: %w", err)
	}
	if len(hosts) == 0 {
		return nil, fmt.Errorf("host not found: %s", ip)
	}
	return &hosts[0], nil
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

	// Separate fields that must be matched after the vuln $lookup (computed fields)
	// from fields that can be matched directly on the hosts collection.
	// rawLookupFilter: matched on the raw _vulns array between $lookup and $addFields.
	// postFilter: matched on computed vuln summary fields after $addFields.
	preFilter := bson.M{}
	rawLookupFilter := bson.M{}
	postFilter := bson.M{}
	for k, v := range filter {
		switch k {
		case "vulns.max_severity":
			postFilter[k] = v
		case "vulns.template_id":
			rawLookupFilter["_vulns.template_id"] = v
		default:
			preFilter[k] = v
		}
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: preFilter}},
	}

	// Convert the hex string to primitive.ObjectID so MongoDB performs a
	// same-type comparison (ObjectID vs ObjectID); cross-type $gt returns no results.
	if lastID != "" {
		if oid, err := primitive.ObjectIDFromHex(lastID); err == nil {
			pipeline = append(pipeline, bson.D{
				{Key: "$match", Value: bson.M{"_id": bson.M{"$gt": oid}}},
			})
		}
	}

	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: bson.M{"_id": 1}}})

	// When filtering by computed vuln fields, limit must come after the lookup.
	// Otherwise limit before lookup for performance (avoids lookup on skipped docs).
	hasPostLookupFilter := len(postFilter) > 0 || len(rawLookupFilter) > 0
	if !hasPostLookupFilter {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: limit + 1}})
	}

	// Insert raw-lookup filter between $lookup and $addFields/$project so we
	// can match on the joined _vulns array before it is projected away.
	lookupStages := repo.vulnLookupStages()
	if len(rawLookupFilter) > 0 {
		pipeline = append(pipeline, lookupStages[0]) // $lookup
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: rawLookupFilter}})
		pipeline = append(pipeline, lookupStages[1:]...) // $addFields, $project
	} else {
		pipeline = append(pipeline, lookupStages...)
	}

	if hasPostLookupFilter && len(postFilter) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: postFilter}})
	}
	if hasPostLookupFilter {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: limit + 1}})
	}

	cursor, err := repo.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, "", fmt.Errorf("mongodb: search aggregation failed: %w", err)
	}
	defer cursor.Close(ctx)

	hosts := []types.Host{}
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
					bson.M{"$match": bson.M{"location.country_code": bson.M{"$exists": true, "$ne": nil}}},
					bson.M{"$group": bson.M{
						"_id": bson.M{
							"code": "$location.country_code",
							"name": "$location.country_name",
						},
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
			Countries: []storage.CountryFacet{},
			ASNs:      []storage.ASNFacet{},
		}, nil
	}

	facetResult := results[0]
	counts := &storage.FacetCounts{
		Services:  make(map[string]int),
		Countries: []storage.CountryFacet{},
		ASNs:      []storage.ASNFacet{},
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
							// Extract code and name from the grouped _id
							if idMap, ok := id.(bson.M); ok {
								code := fmt.Sprintf("%v", idMap["code"])
								name := fmt.Sprintf("%v", idMap["name"])
								counts.Countries = append(counts.Countries, storage.CountryFacet{
									Code:  code,
									Name:  name,
									Count: int(count),
								})
							}
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
							if idMap, ok := id.(bson.M); ok {
								var asnCode uint32
								if numberVal, ok := idMap["number"]; ok {
									// Handle both int32 and uint32 types
									switch v := numberVal.(type) {
									case int32:
										asnCode = uint32(v)
									case uint32:
										asnCode = v
									case int64:
										asnCode = uint32(v)
									case float64:
										asnCode = uint32(v)
									}
								}

								var organization string
								if orgVal, ok := idMap["organization"]; ok {
									organization = fmt.Sprintf("%v", orgVal)
								}

								counts.ASNs = append(counts.ASNs, storage.ASNFacet{
									Code:  asnCode,
									Name:  organization,
									Count: int(count),
								})
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
