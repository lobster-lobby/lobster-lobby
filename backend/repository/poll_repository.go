package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/lobster-lobby/lobster-lobby/models"
)

type PollRepository struct {
	polls *mongo.Collection
	votes *mongo.Collection
}

func NewPollRepository(db *MongoDB) *PollRepository {
	return &PollRepository{
		polls: db.Database.Collection("polls"),
		votes: db.Database.Collection("poll_votes"),
	}
}

func (r *PollRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.polls.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "policyId", Value: 1}, {Key: "createdAt", Value: -1}}},
	})
	if err != nil {
		return err
	}
	_, err = r.votes.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "pollId", Value: 1}, {Key: "userId", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	return err
}

func (r *PollRepository) Create(ctx context.Context, poll *models.Poll) (*models.Poll, error) {
	now := time.Now().UTC()
	poll.CreatedAt = now
	poll.UpdatedAt = now
	if poll.ID.IsZero() {
		poll.ID = bson.NewObjectID()
	}
	if poll.Status == "" {
		poll.Status = "active"
	}
	for i := range poll.Options {
		if poll.Options[i].ID.IsZero() {
			poll.Options[i].ID = bson.NewObjectID()
		}
	}
	_, err := r.polls.InsertOne(ctx, poll)
	if err != nil {
		return nil, err
	}
	return poll, nil
}

func (r *PollRepository) ListByPolicy(ctx context.Context, policyID bson.ObjectID) ([]models.Poll, error) {
	cursor, err := r.polls.Find(ctx,
		bson.M{"policyId": policyID},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var polls []models.Poll
	if err := cursor.All(ctx, &polls); err != nil {
		return nil, err
	}
	if polls == nil {
		polls = []models.Poll{}
	}
	return polls, nil
}

func (r *PollRepository) GetByID(ctx context.Context, id bson.ObjectID) (*models.Poll, error) {
	var poll models.Poll
	err := r.polls.FindOne(ctx, bson.M{"_id": id}).Decode(&poll)
	if err != nil {
		return nil, err
	}
	return &poll, nil
}

func (r *PollRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	_, err := r.polls.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	_, _ = r.votes.DeleteMany(ctx, bson.M{"pollId": id})
	return nil
}

// Vote casts or replaces a user's vote on a poll (idempotent).
// Vote counts are updated atomically via $inc to prevent concurrent write corruption.
func (r *PollRepository) Vote(ctx context.Context, pollID, userID bson.ObjectID, optionIDs []bson.ObjectID) (*models.Poll, error) {
	// Get the existing vote if any
	var existing models.PollVote
	existErr := r.votes.FindOne(ctx, bson.M{"pollId": pollID, "userId": userID}).Decode(&existing)
	if existErr != nil && !errors.Is(existErr, mongo.ErrNoDocuments) {
		return nil, existErr
	}
	hasExisting := existErr == nil

	// Atomically decrement counts for old vote options
	if hasExisting {
		for _, optID := range existing.OptionIDs {
			_, err := r.polls.UpdateOne(ctx,
				bson.M{"_id": pollID, "options.id": optID},
				bson.M{"$inc": bson.M{"options.$.votes": -1}},
			)
			if err != nil {
				return nil, err
			}
		}
		_, err := r.polls.UpdateOne(ctx,
			bson.M{"_id": pollID},
			bson.M{"$inc": bson.M{"totalVotes": -1}},
		)
		if err != nil {
			return nil, err
		}
	}

	// Atomically increment counts for new vote options
	for _, optID := range optionIDs {
		_, err := r.polls.UpdateOne(ctx,
			bson.M{"_id": pollID, "options.id": optID},
			bson.M{"$inc": bson.M{"options.$.votes": 1}},
		)
		if err != nil {
			return nil, err
		}
	}
	_, err := r.polls.UpdateOne(ctx,
		bson.M{"_id": pollID},
		bson.M{
			"$inc": bson.M{"totalVotes": 1},
			"$set": bson.M{"updatedAt": time.Now().UTC()},
		},
	)
	if err != nil {
		return nil, err
	}

	// Upsert vote record
	_, err = r.votes.UpdateOne(ctx,
		bson.M{"pollId": pollID, "userId": userID},
		bson.M{"$set": bson.M{
			"pollId":    pollID,
			"userId":    userID,
			"optionIds": optionIDs,
			"createdAt": time.Now().UTC(),
		}},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, pollID)
}

// GetUserVote returns the user's current vote on a poll (nil if none).
func (r *PollRepository) GetUserVote(ctx context.Context, pollID, userID bson.ObjectID) *models.PollVote {
	var vote models.PollVote
	err := r.votes.FindOne(ctx, bson.M{"pollId": pollID, "userId": userID}).Decode(&vote)
	if err != nil {
		return nil
	}
	return &vote
}
