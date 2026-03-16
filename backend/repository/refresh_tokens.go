package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type RefreshToken struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	Token     string        `bson:"token"`
	UserID    bson.ObjectID `bson:"userID"`
	ExpiresAt time.Time     `bson:"expiresAt"`
	CreatedAt time.Time     `bson:"createdAt"`
}

type RefreshTokenRepository struct {
	coll *mongo.Collection
}

func NewRefreshTokenRepository(db *MongoDB) *RefreshTokenRepository {
	return &RefreshTokenRepository{coll: db.Database.Collection("refresh_tokens")}
}

func (r *RefreshTokenRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "token", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "expiresAt", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
	})
	return err
}

func (r *RefreshTokenRepository) Create(ctx context.Context, rt *RefreshToken) error {
	rt.CreatedAt = time.Now().UTC()
	if rt.ID.IsZero() {
		rt.ID = bson.NewObjectID()
	}
	_, err := r.coll.InsertOne(ctx, rt)
	return err
}

func (r *RefreshTokenRepository) FindByToken(ctx context.Context, token string) (*RefreshToken, error) {
	var rt RefreshToken
	err := r.coll.FindOne(ctx, bson.M{"token": token}).Decode(&rt)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &rt, err
}

func (r *RefreshTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"token": token})
	return err
}

func (r *RefreshTokenRepository) DeleteAllForUser(ctx context.Context, userID bson.ObjectID) error {
	_, err := r.coll.DeleteMany(ctx, bson.M{"userID": userID})
	return err
}
