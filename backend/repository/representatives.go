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
		{Keys: bson.D{{Key: "state", Value: 1}}},
		{Keys: bson.D{{Key: "district", Value: 1}}},
		{Keys: bson.D{{Key: "state", Value: 1}, {Key: "chamber", Value: 1}}},
		{Keys: bson.D{{Key: "party", Value: 1}}},
		{Keys: bson.D{{Key: "chamber", Value: 1}}},
		{Keys: bson.D{{Key: "name", Value: 1}}},
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

// RepListOpts configures pagination and filtering for listing representatives
type RepListOpts struct {
	Page    int
	PerPage int
	Search  string
	Party   string
	State   string
	Chamber string
}

func (r *RepresentativeRepository) Create(ctx context.Context, rep *models.Representative) error {
	now := time.Now().UTC()
	rep.CreatedAt = now
	rep.UpdatedAt = now
	if rep.ID.IsZero() {
		rep.ID = bson.NewObjectID()
	}
	_, err := r.coll.InsertOne(ctx, rep)
	return err
}

func (r *RepresentativeRepository) Update(ctx context.Context, id bson.ObjectID, updates bson.M) error {
	updates["updatedAt"] = time.Now().UTC()
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": updates})
	return err
}

func (r *RepresentativeRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *RepresentativeRepository) List(ctx context.Context, opts RepListOpts) ([]models.Representative, int64, error) {
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PerPage < 1 || opts.PerPage > 100 {
		opts.PerPage = 20
	}

	filter := bson.M{}
	if opts.Search != "" {
		filter["name"] = bson.M{"$regex": opts.Search, "$options": "i"}
	}
	if opts.Party != "" {
		filter["party"] = opts.Party
	}
	if opts.State != "" {
		filter["state"] = opts.State
	}
	if opts.Chamber != "" {
		filter["chamber"] = opts.Chamber
	}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpts := options.Find().
		SetSort(bson.M{"name": 1}).
		SetSkip(int64((opts.Page - 1) * opts.PerPage)).
		SetLimit(int64(opts.PerPage))

	cursor, err := r.coll.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var reps []models.Representative
	if err := cursor.All(ctx, &reps); err != nil {
		return nil, 0, err
	}
	if reps == nil {
		reps = []models.Representative{}
	}
	return reps, total, nil
}
