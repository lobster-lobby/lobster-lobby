package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/lobster-lobby/lobster-lobby/models"
)

type ReputationRepository struct {
	coll *mongo.Collection
}

func NewReputationRepository(db *MongoDB) *ReputationRepository {
	return &ReputationRepository{coll: db.Database.Collection("reputation_events")}
}

func (r *ReputationRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "createdAt", Value: -1},
			},
		},
	})
	return err
}

func (r *ReputationRepository) LogEvent(ctx context.Context, event *models.ReputationEvent) error {
	event.CreatedAt = time.Now().UTC()
	if event.ID.IsZero() {
		event.ID = bson.NewObjectID()
	}
	_, err := r.coll.InsertOne(ctx, event)
	return err
}

func (r *ReputationRepository) SumByUser(ctx context.Context, userID bson.ObjectID) (int, int, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{"userId": userID}},
		bson.M{"$group": bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$points"},
			"count": bson.M{"$sum": 1},
		}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, 0, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Total int `bson:"total"`
		Count int `bson:"count"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return 0, 0, err
	}

	if len(results) == 0 {
		return 0, 0, nil
	}
	return results[0].Total, results[0].Count, nil
}

func (r *ReputationRepository) ListByUser(ctx context.Context, userID bson.ObjectID, page, perPage int) ([]models.ReputationEvent, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	filter := bson.M{"userId": userID}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(int64((page - 1) * perPage)).
		SetLimit(int64(perPage))

	cursor, err := r.coll.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var events []models.ReputationEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, 0, err
	}

	if events == nil {
		events = []models.ReputationEvent{}
	}

	return events, total, nil
}
