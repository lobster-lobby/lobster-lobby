package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/lobster-lobby/lobster-lobby/models"
)

type RepresentativeRepository struct {
	coll *mongo.Collection
}

func NewRepresentativeRepository(db *MongoDB) *RepresentativeRepository {
	return &RepresentativeRepository{coll: db.Database.Collection("representatives")}
}

func (r *RepresentativeRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "state", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "district", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "state", Value: 1},
				{Key: "chamber", Value: 1},
			},
		},
	})
	return err
}

func (r *RepresentativeRepository) FindByID(ctx context.Context, id bson.ObjectID) (*models.Representative, error) {
	var rep models.Representative
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&rep)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &rep, err
}

func (r *RepresentativeRepository) FindByState(ctx context.Context, state string) ([]models.Representative, error) {
	cursor, err := r.coll.Find(ctx, bson.M{"state": state})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reps []models.Representative
	if err := cursor.All(ctx, &reps); err != nil {
		return nil, err
	}
	if reps == nil {
		reps = []models.Representative{}
	}
	return reps, nil
}

func (r *RepresentativeRepository) FindByDistrict(ctx context.Context, district string) ([]models.Representative, error) {
	cursor, err := r.coll.Find(ctx, bson.M{"district": district})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reps []models.Representative
	if err := cursor.All(ctx, &reps); err != nil {
		return nil, err
	}
	if reps == nil {
		reps = []models.Representative{}
	}
	return reps, nil
}

func (r *RepresentativeRepository) Upsert(ctx context.Context, rep *models.Representative) error {
	now := time.Now().UTC()
	rep.UpdatedAt = now

	filter := bson.M{
		"name":   rep.Name,
		"state":  rep.State,
		"chamber": rep.Chamber,
	}

	update := bson.M{
		"$set": rep,
		"$setOnInsert": bson.M{
			"createdAt": now,
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.coll.UpdateOne(ctx, filter, update, opts)
	return err
}
