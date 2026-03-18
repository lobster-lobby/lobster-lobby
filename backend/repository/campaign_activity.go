package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/lobster-lobby/lobster-lobby/models"
)

type CampaignActivityRepository struct {
	coll *mongo.Collection
}

func NewCampaignActivityRepository(db *MongoDB) *CampaignActivityRepository {
	return &CampaignActivityRepository{coll: db.Database.Collection("campaign_activities")}
}

func (r *CampaignActivityRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "campaignId", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "userId", Value: 1}}},
		{Keys: bson.D{{Key: "type", Value: 1}}},
	})
	return err
}

func (r *CampaignActivityRepository) Create(ctx context.Context, activity *models.CampaignActivity) error {
	activity.CreatedAt = time.Now().UTC()
	if activity.ID.IsZero() {
		activity.ID = bson.NewObjectID()
	}
	_, err := r.coll.InsertOne(ctx, activity)
	return err
}

type CampaignActivityListOpts struct {
	CampaignID string
	Page       int
	Limit      int
}

func (r *CampaignActivityRepository) ListByCampaign(ctx context.Context, opts CampaignActivityListOpts) ([]models.CampaignActivity, int64, error) {
	campaignOID, err := bson.ObjectIDFromHex(opts.CampaignID)
	if err != nil {
		return []models.CampaignActivity{}, 0, nil
	}

	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.Limit < 1 || opts.Limit > 100 {
		opts.Limit = 20
	}

	filter := bson.M{"campaignId": campaignOID}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(int64((opts.Page - 1) * opts.Limit)).
		SetLimit(int64(opts.Limit))

	cursor, err := r.coll.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var activities []models.CampaignActivity
	if err := cursor.All(ctx, &activities); err != nil {
		return nil, 0, err
	}

	if activities == nil {
		activities = []models.CampaignActivity{}
	}
	return activities, total, nil
}

// CountUniqueUsers returns the number of unique users who have activity on a campaign.
func (r *CampaignActivityRepository) CountUniqueUsers(ctx context.Context, campaignID string) (int, error) {
	campaignOID, err := bson.ObjectIDFromHex(campaignID)
	if err != nil {
		return 0, nil
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"campaignId": campaignOID}}},
		{{Key: "$group", Value: bson.M{"_id": "$userId"}}},
		{{Key: "$count", Value: "count"}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result []struct {
		Count int `bson:"count"`
	}
	if err := cursor.All(ctx, &result); err != nil {
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil
	}
	return result[0].Count, nil
}
