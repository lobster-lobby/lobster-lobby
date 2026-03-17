package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/lobster-lobby/lobster-lobby/models"
)

type CampaignEventRepository struct {
	coll *mongo.Collection
}

func NewCampaignEventRepository(db *MongoDB) *CampaignEventRepository {
	return &CampaignEventRepository{coll: db.Database.Collection("campaign_events")}
}

func (r *CampaignEventRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "campaignId", Value: 1}}},
		{Keys: bson.D{{Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "type", Value: 1}}},
	})
	return err
}

func (r *CampaignEventRepository) Create(ctx context.Context, event *models.CampaignEvent) error {
	event.CreatedAt = time.Now().UTC()
	if event.ID.IsZero() {
		event.ID = bson.NewObjectID()
	}
	if event.Metadata == nil {
		event.Metadata = make(map[string]any)
	}

	_, err := r.coll.InsertOne(ctx, event)
	return err
}

func (r *CampaignEventRepository) ListByCampaign(ctx context.Context, campaignID string) ([]models.CampaignEvent, error) {
	campaignOID, err := bson.ObjectIDFromHex(campaignID)
	if err != nil {
		return []models.CampaignEvent{}, nil
	}

	filter := bson.M{"campaignId": campaignOID}
	findOpts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.coll.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []models.CampaignEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	if events == nil {
		events = []models.CampaignEvent{}
	}
	return events, nil
}

// DailyActivity represents activity counts for a single day.
type DailyActivity struct {
	Date  string `bson:"_id" json:"date"`
	Count int    `bson:"count" json:"count"`
}

// GetActivityByDay returns daily event counts for the last N days.
func (r *CampaignEventRepository) GetActivityByDay(ctx context.Context, campaignID string, days int) ([]DailyActivity, error) {
	campaignOID, err := bson.ObjectIDFromHex(campaignID)
	if err != nil {
		return []DailyActivity{}, nil
	}

	startDate := time.Now().UTC().AddDate(0, 0, -days)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"campaignId": campaignOID,
			"createdAt":  bson.M{"$gte": startDate},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{
					"format": "%Y-%m-%d",
					"date":   "$createdAt",
				},
			},
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"_id": 1}}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []DailyActivity
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if results == nil {
		results = []DailyActivity{}
	}
	return results, nil
}

// GetRecentEvents returns the N most recent events for a campaign.
func (r *CampaignEventRepository) GetRecentEvents(ctx context.Context, campaignID string, limit int) ([]models.CampaignEvent, error) {
	campaignOID, err := bson.ObjectIDFromHex(campaignID)
	if err != nil {
		return []models.CampaignEvent{}, nil
	}

	filter := bson.M{"campaignId": campaignOID}
	findOpts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.coll.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []models.CampaignEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	if events == nil {
		events = []models.CampaignEvent{}
	}
	return events, nil
}
