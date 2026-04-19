package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/ctrlsam/rigour/internal/storage"
	"github.com/ctrlsam/rigour/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type VulnerabilityRepository struct {
	collection *mongo.Collection
}

func NewVulnerabilityRepository(ctx context.Context, coll *mongo.Collection) (storage.VulnerabilityRepository, error) {
	_, _ = coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "ip", Value: 1},
				{Key: "port", Value: 1},
				{Key: "template_id", Value: 1},
			},
			Options: options.Index().SetName("ip_port_template_unique").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "ip", Value: 1}},
			Options: options.Index().SetName("ip_idx"),
		},
		{
			Keys:    bson.D{{Key: "campaign_id", Value: 1}},
			Options: options.Index().SetName("campaign_id_idx"),
		},
		{
			Keys:    bson.D{{Key: "severity", Value: 1}},
			Options: options.Index().SetName("severity_idx"),
		},
	})
	return &VulnerabilityRepository{collection: coll}, nil
}

func (r *VulnerabilityRepository) UpsertVulnerability(ctx context.Context, v types.Vulnerability) error {
	filter := bson.M{
		"ip":          v.IP,
		"port":        v.Port,
		"template_id": v.TemplateID,
	}

	if v.ID == "" {
		v.ID = primitive.NewObjectID().Hex()
	}
	if v.DiscoveredAt.IsZero() {
		v.DiscoveredAt = time.Now().UTC()
	}

	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":          v.ID,
			"discovered_at": v.DiscoveredAt,
		},
		"$set": bson.M{
			"ip":          v.IP,
			"port":        v.Port,
			"protocol":    v.Protocol,
			"template_id": v.TemplateID,
			"name":        v.Name,
			"severity":    v.Severity,
			"description": v.Description,
			"cvss":        v.CVSS,
			"cve_ids":     v.CVEIDs,
			"matched_at":  v.MatchedAt,
			"campaign_id": v.CampaignID,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("upsert vulnerability: %w", err)
	}
	return nil
}

func (r *VulnerabilityRepository) ListByIP(ctx context.Context, ip string) ([]types.Vulnerability, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"ip": ip})
	if err != nil {
		return nil, fmt.Errorf("list vulnerabilities by ip: %w", err)
	}
	defer cursor.Close(ctx)

	results := make([]types.Vulnerability, 0)
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode vulnerabilities: %w", err)
	}
	return results, nil
}

func (r *VulnerabilityRepository) ListAll(ctx context.Context, severity string) ([]types.Vulnerability, error) {
	matchStage := bson.M{}
	if severity != "" {
		matchStage["severity"] = severity
	}

	severityOrder := bson.M{"$switch": bson.M{
		"branches": bson.A{
			bson.M{"case": bson.M{"$eq": bson.A{"$severity", "critical"}}, "then": 0},
			bson.M{"case": bson.M{"$eq": bson.A{"$severity", "high"}}, "then": 1},
			bson.M{"case": bson.M{"$eq": bson.A{"$severity", "medium"}}, "then": 2},
			bson.M{"case": bson.M{"$eq": bson.A{"$severity", "low"}}, "then": 3},
		},
		"default": 4,
	}}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$addFields", Value: bson.M{"_severity_order": severityOrder}}},
		{{Key: "$sort", Value: bson.D{{Key: "_severity_order", Value: 1}, {Key: "cvss", Value: -1}}}},
		{{Key: "$project", Value: bson.M{"_severity_order": 0}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("list all vulnerabilities: %w", err)
	}
	defer cursor.Close(ctx)

	results := make([]types.Vulnerability, 0)
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode vulnerabilities: %w", err)
	}
	return results, nil
}

func (r *VulnerabilityRepository) SummariseByTemplate(ctx context.Context, severity string) ([]types.VulnerabilityTemplateSummary, error) {
	matchStage := bson.M{}
	if severity != "" {
		matchStage["severity"] = severity
	}

	severityOrder := bson.M{"$switch": bson.M{
		"branches": bson.A{
			bson.M{"case": bson.M{"$eq": bson.A{"$severity", "critical"}}, "then": 0},
			bson.M{"case": bson.M{"$eq": bson.A{"$severity", "high"}}, "then": 1},
			bson.M{"case": bson.M{"$eq": bson.A{"$severity", "medium"}}, "then": 2},
			bson.M{"case": bson.M{"$eq": bson.A{"$severity", "low"}}, "then": 3},
		},
		"default": 4,
	}}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$group", Value: bson.M{
			"_id":         "$template_id",
			"name":        bson.M{"$first": "$name"},
			"severity":    bson.M{"$first": "$severity"},
			"cvss":        bson.M{"$max": "$cvss"},
			"description": bson.M{"$first": "$description"},
			"_cve_sets":   bson.M{"$addToSet": "$cve_ids"},
			"_ips":        bson.M{"$addToSet": "$ip"},
		}}},
		{{Key: "$addFields", Value: bson.M{
			"affected_host_count": bson.M{"$size": "$_ips"},
			"cve_ids": bson.M{"$reduce": bson.M{
				"input":        "$_cve_sets",
				"initialValue": bson.A{},
				"in":           bson.M{"$setUnion": bson.A{"$$value", "$$this"}},
			}},
			"_severity_order": severityOrder,
		}}},
		{{Key: "$sort", Value: bson.D{
			{Key: "_severity_order", Value: 1},
			{Key: "affected_host_count", Value: -1},
			{Key: "cvss", Value: -1},
		}}},
		{{Key: "$project", Value: bson.M{"_severity_order": 0, "_cve_sets": 0, "_ips": 0}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("summarise vulnerabilities by template: %w", err)
	}
	defer cursor.Close(ctx)

	results := make([]types.VulnerabilityTemplateSummary, 0)
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode vulnerability summaries: %w", err)
	}
	return results, nil
}

func (r *VulnerabilityRepository) ListByCampaign(ctx context.Context, campaignID string, severity string) ([]types.Vulnerability, error) {
	filter := bson.M{"campaign_id": campaignID}
	if severity != "" {
		filter["severity"] = severity
	}

	cursor, err := r.collection.Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "severity", Value: 1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("list vulnerabilities by campaign: %w", err)
	}
	defer cursor.Close(ctx)

	results := make([]types.Vulnerability, 0)
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode vulnerabilities: %w", err)
	}
	return results, nil
}
