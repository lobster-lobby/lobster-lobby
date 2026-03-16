package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/lobster-lobby/lobster-lobby/models"
)

type APIKeyRepository struct {
	coll *mongo.Collection
}

func NewAPIKeyRepository(db *MongoDB) *APIKeyRepository {
	return &APIKeyRepository{coll: db.Database.Collection("api_keys")}
}

func (r *APIKeyRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "keyPrefix", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "userId", Value: 1}},
		},
	})
	return err
}

func (r *APIKeyRepository) Create(ctx context.Context, key *models.APIKey) error {
	key.CreatedAt = time.Now().UTC()
	if key.ID.IsZero() {
		key.ID = bson.NewObjectID()
	}
	_, err := r.coll.InsertOne(ctx, key)
	return err
}

func (r *APIKeyRepository) FindByPrefix(ctx context.Context, prefix string) (*models.APIKey, error) {
	var key models.APIKey
	err := r.coll.FindOne(ctx, bson.M{"keyPrefix": prefix}).Decode(&key)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &key, err
}

func (r *APIKeyRepository) FindByUserID(ctx context.Context, userID bson.ObjectID) ([]models.APIKey, error) {
	cursor, err := r.coll.Find(ctx, bson.M{"userId": userID, "revoked": false})
	if err != nil {
		return nil, err
	}
	var keys []models.APIKey
	if err := cursor.All(ctx, &keys); err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, id bson.ObjectID) error {
	now := time.Now().UTC()
	_, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"lastUsedAt": now}},
	)
	return err
}

func (r *APIKeyRepository) Revoke(ctx context.Context, id, userID bson.ObjectID) error {
	result, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": id, "userId": userID},
		bson.M{"$set": bson.M{"revoked": true}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *APIKeyRepository) Delete(ctx context.Context, id, userID bson.ObjectID) error {
	result, err := r.coll.DeleteOne(ctx, bson.M{"_id": id, "userId": userID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
