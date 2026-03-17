package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/lobster-lobby/lobster-lobby/models"
)

type NominationRepository struct {
	coll     *mongo.Collection
	comments *mongo.Collection
	research *mongo.Collection
}

func NewNominationRepository(db *MongoDB) *NominationRepository {
	return &NominationRepository{
		coll:     db.Database.Collection("nominations"),
		comments: db.Database.Collection("comments"),
		research: db.Database.Collection("research"),
	}
}

func (r *NominationRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "policyId", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "nominatedBy", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
	})
	return err
}

func (r *NominationRepository) Create(ctx context.Context, nomination *models.CampaignNomination) error {
	now := time.Now().UTC()
	nomination.CreatedAt = now
	nomination.UpdatedAt = now
	if nomination.ID.IsZero() {
		nomination.ID = bson.NewObjectID()
	}
	if nomination.Endorsers == nil {
		nomination.Endorsers = []models.NominationEndorsement{}
	}
	_, err := r.coll.InsertOne(ctx, nomination)
	return err
}

func (r *NominationRepository) FindByPolicyID(ctx context.Context, policyID bson.ObjectID) (*models.CampaignNomination, error) {
	var n models.CampaignNomination
	err := r.coll.FindOne(ctx, bson.M{"policyId": policyID}).Decode(&n)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &n, err
}

func (r *NominationRepository) AddEndorsement(ctx context.Context, policyID, userID bson.ObjectID) (*models.CampaignNomination, error) {
	now := time.Now().UTC()
	endorsement := models.NominationEndorsement{
		UserID:    userID,
		CreatedAt: now,
	}

	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{
			"policyId":          policyID,
			"status":            models.NominationStatusPending,
			"endorsers.userId":  bson.M{"$ne": userID},
		},
		bson.M{
			"$push": bson.M{"endorsers": endorsement},
			"$set":  bson.M{"updatedAt": now},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var n models.CampaignNomination
	if err := result.Decode(&n); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &n, nil
}

func (r *NominationRepository) UpdateStatus(ctx context.Context, policyID bson.ObjectID, status models.NominationStatus) error {
	_, err := r.coll.UpdateOne(
		ctx,
		bson.M{"policyId": policyID},
		bson.M{"$set": bson.M{"status": status, "updatedAt": time.Now().UTC()}},
	)
	return err
}

// CountDebateComments returns the total number of debate comments for a policy.
func (r *NominationRepository) CountDebateComments(ctx context.Context, policyID bson.ObjectID) (int64, error) {
	return r.comments.CountDocuments(ctx, bson.M{"policyId": policyID})
}

// CountResearchSubmissions returns the total number of research submissions for a policy.
func (r *NominationRepository) CountResearchSubmissions(ctx context.Context, policyID bson.ObjectID) (int64, error) {
	return r.research.CountDocuments(ctx, bson.M{"policyId": policyID})
}
