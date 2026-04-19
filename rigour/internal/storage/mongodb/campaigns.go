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

type CampaignRepository struct {
	collection *mongo.Collection
}

func NewCampaignRepository(ctx context.Context, coll *mongo.Collection) (storage.CampaignRepository, error) {
	if _, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "status", Value: 1}},
			Options: options.Index().SetName("status_idx"),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("created_at_idx"),
		},
	}); err != nil {
		slog.Warn("failed to create campaigns indexes", "collection", coll.Name(), "err", err)
	}
	return &CampaignRepository{collection: coll}, nil
}

func (r *CampaignRepository) CreateCampaign(ctx context.Context, c *types.Campaign) (*types.Campaign, error) {
	now := time.Now().UTC()
	c.ID = primitive.NewObjectID().Hex()
	c.Status = types.CampaignPending
	c.CreatedAt = now
	c.UpdatedAt = now

	if _, err := r.collection.InsertOne(ctx, c); err != nil {
		return nil, fmt.Errorf("insert campaign: %w", err)
	}
	return c, nil
}

func (r *CampaignRepository) GetCampaign(ctx context.Context, id string) (*types.Campaign, error) {
	var c types.Campaign
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&c)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("campaign %q not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get campaign: %w", err)
	}
	return &c, nil
}

func (r *CampaignRepository) ListCampaigns(ctx context.Context, filter storage.CampaignFilter) ([]types.Campaign, error) {
	query := bson.M{}
	if filter.Status != "" {
		query["status"] = filter.Status
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cur, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("list campaigns: %w", err)
	}
	defer cur.Close(ctx)

	var campaigns []types.Campaign
	if err := cur.All(ctx, &campaigns); err != nil {
		return nil, fmt.Errorf("decode campaigns: %w", err)
	}
	return campaigns, nil
}

func (r *CampaignRepository) UpdateCampaignStatus(ctx context.Context, id string, status types.CampaignStatus, errMsg string) error {
	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": now,
			"error":      errMsg,
		},
	}

	switch status {
	case types.CampaignRunning:
		update["$set"].(bson.M)["started_at"] = now
	case types.CampaignCompleted, types.CampaignFailed, types.CampaignCancelled:
		update["$set"].(bson.M)["completed_at"] = now
	}

	res, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("update campaign status: %w", err)
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("campaign %q not found", id)
	}
	return nil
}

func (r *CampaignRepository) UpdateCampaignProgress(ctx context.Context, id string, progress types.CampaignProgress) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"progress":   progress,
			"updated_at": time.Now().UTC(),
		},
	})
	if err != nil {
		return fmt.Errorf("update campaign progress: %w", err)
	}
	return nil
}

func (r *CampaignRepository) SetCampaignCIDRs(ctx context.Context, id string, cidrs []string) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"cidrs":      cidrs,
			"updated_at": time.Now().UTC(),
		},
	})
	if err != nil {
		return fmt.Errorf("set campaign cidrs: %w", err)
	}
	return nil
}

func (r *CampaignRepository) DeleteCampaign(ctx context.Context, id string) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete campaign: %w", err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("campaign %q not found", id)
	}
	return nil
}
