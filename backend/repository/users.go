package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/lobster-lobby/lobster-lobby/models"
)

type UserRepository struct {
	coll *mongo.Collection
}

func NewUserRepository(db *MongoDB) *UserRepository {
	return &UserRepository{coll: db.Database.Collection("users")}
}

func (r *UserRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
	})
	return err
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	if user.ID.IsZero() {
		user.ID = bson.NewObjectID()
	}
	_, err := r.coll.InsertOne(ctx, user)
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &u, err
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var u models.User
	err := r.coll.FindOne(ctx, bson.M{"username": username}).Decode(&u)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &u, err
}

func (r *UserRepository) FindByID(ctx context.Context, id bson.ObjectID) (*models.User, error) {
	var u models.User
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &u, err
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id bson.ObjectID) error {
	now := time.Now().UTC()
	_, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"lastLoginAt": now, "updatedAt": now}},
	)
	return err
}

func (r *UserRepository) UpdateReputation(ctx context.Context, id bson.ObjectID, rep models.ReputationScore) error {
	_, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"reputation": rep, "updatedAt": time.Now().UTC()}},
	)
	return err
}

func (r *UserRepository) Update(ctx context.Context, id bson.ObjectID, username, email, displayName, bio string) error {
	update := bson.M{"updatedAt": time.Now().UTC()}
	if username != "" {
		update["username"] = username
	}
	if email != "" {
		update["email"] = email
	}
	if displayName != "" {
		update["displayName"] = displayName
	}
	update["bio"] = bio // Allow empty bio

	_, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	)
	return err
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id bson.ObjectID, newPasswordHash string) error {
	_, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"passwordHash": newPasswordHash, "updatedAt": time.Now().UTC()}},
	)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
