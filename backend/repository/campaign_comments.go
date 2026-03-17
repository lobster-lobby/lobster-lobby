package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/lobster-lobby/lobster-lobby/models"
)

type CampaignCommentRepository struct {
	coll     *mongo.Collection
	voteColl *mongo.Collection
}

func NewCampaignCommentRepository(db *MongoDB) *CampaignCommentRepository {
	return &CampaignCommentRepository{
		coll:     db.Database.Collection("campaign_comments"),
		voteColl: db.Database.Collection("campaign_comment_votes"),
	}
}

func (r *CampaignCommentRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "campaignId", Value: 1}}},
		{Keys: bson.D{{Key: "parentId", Value: 1}}},
		{Keys: bson.D{{Key: "authorId", Value: 1}}},
		{Keys: bson.D{{Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "votes", Value: -1}}},
	})
	if err != nil {
		return err
	}

	_, err = r.voteColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "commentId", Value: 1}, {Key: "userId", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	return err
}

type CampaignCommentListOpts struct {
	CampaignID string
	Sort       string // "newest", "votes"
}

func (r *CampaignCommentRepository) Create(ctx context.Context, comment *models.CampaignComment) error {
	now := time.Now().UTC()
	comment.CreatedAt = now
	comment.UpdatedAt = now
	if comment.ID.IsZero() {
		comment.ID = bson.NewObjectID()
	}
	comment.Votes = 0

	_, err := r.coll.InsertOne(ctx, comment)
	return err
}

func (r *CampaignCommentRepository) GetByID(ctx context.Context, id string) (*models.CampaignComment, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var comment models.CampaignComment
	err = r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&comment)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &comment, err
}

func (r *CampaignCommentRepository) ListByCampaign(ctx context.Context, opts CampaignCommentListOpts) ([]models.CampaignComment, error) {
	campaignOID, err := bson.ObjectIDFromHex(opts.CampaignID)
	if err != nil {
		return []models.CampaignComment{}, nil
	}

	filter := bson.M{"campaignId": campaignOID}

	sortField := r.getSortField(opts.Sort)
	findOpts := options.Find().SetSort(sortField)

	cursor, err := r.coll.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comments []models.CampaignComment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}

	if comments == nil {
		comments = []models.CampaignComment{}
	}
	return comments, nil
}

func (r *CampaignCommentRepository) Update(ctx context.Context, id string, updates bson.M) (*models.CampaignComment, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	updates["updatedAt"] = time.Now().UTC()
	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"_id": oid},
		bson.M{"$set": updates},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var comment models.CampaignComment
	if err := result.Decode(&comment); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &comment, nil
}

func (r *CampaignCommentRepository) Delete(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Delete the comment
	_, err = r.coll.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return err
	}

	// Delete associated votes
	_, err = r.voteColl.DeleteMany(ctx, bson.M{"commentId": oid})
	return err
}

// ToggleVote sets or removes a user's vote on a comment.
// If the user already voted with the same value, the vote is removed.
// Returns the new vote value (0 if removed).
func (r *CampaignCommentRepository) ToggleVote(ctx context.Context, commentID, userID string, value int) (int, error) {
	commentOID, err := bson.ObjectIDFromHex(commentID)
	if err != nil {
		return 0, err
	}
	userOID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}

	// Check existing vote
	var existingVote models.CampaignCommentVote
	err = r.voteColl.FindOne(ctx, bson.M{
		"commentId": commentOID,
		"userId":    userOID,
	}).Decode(&existingVote)

	now := time.Now().UTC()

	if err == mongo.ErrNoDocuments {
		// No existing vote - create new one
		vote := models.CampaignCommentVote{
			ID:        bson.NewObjectID(),
			CommentID: commentOID,
			UserID:    userOID,
			Value:     value,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if _, err := r.voteColl.InsertOne(ctx, vote); err != nil {
			return 0, err
		}
		// Update comment vote count
		if _, err := r.coll.UpdateOne(ctx, bson.M{"_id": commentOID}, bson.M{"$inc": bson.M{"votes": value}}); err != nil {
			return 0, err
		}
		return value, nil
	} else if err != nil {
		return 0, err
	}

	// Existing vote found
	if existingVote.Value == value {
		// Same vote - remove it (toggle off)
		if _, err := r.voteColl.DeleteOne(ctx, bson.M{"_id": existingVote.ID}); err != nil {
			return 0, err
		}
		// Decrement vote count
		if _, err := r.coll.UpdateOne(ctx, bson.M{"_id": commentOID}, bson.M{"$inc": bson.M{"votes": -value}}); err != nil {
			return 0, err
		}
		return 0, nil
	}

	// Different vote - update it
	delta := value - existingVote.Value
	if _, err := r.voteColl.UpdateOne(ctx, bson.M{"_id": existingVote.ID}, bson.M{
		"$set": bson.M{"value": value, "updatedAt": now},
	}); err != nil {
		return 0, err
	}
	// Update comment vote count
	if _, err := r.coll.UpdateOne(ctx, bson.M{"_id": commentOID}, bson.M{"$inc": bson.M{"votes": delta}}); err != nil {
		return 0, err
	}
	return value, nil
}

// GetUserVote returns the user's vote on a comment (0 if not voted).
func (r *CampaignCommentRepository) GetUserVote(ctx context.Context, commentID, userID string) (int, error) {
	commentOID, err := bson.ObjectIDFromHex(commentID)
	if err != nil {
		return 0, nil
	}
	userOID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return 0, nil
	}

	var vote models.CampaignCommentVote
	err = r.voteColl.FindOne(ctx, bson.M{
		"commentId": commentOID,
		"userId":    userOID,
	}).Decode(&vote)

	if err == mongo.ErrNoDocuments {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return vote.Value, nil
}

// GetBatchUserVotes returns user votes for multiple comments.
func (r *CampaignCommentRepository) GetBatchUserVotes(ctx context.Context, commentIDs []string, userID string) (map[string]int, error) {
	userOID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return map[string]int{}, nil
	}

	oids := make([]bson.ObjectID, 0, len(commentIDs))
	for _, id := range commentIDs {
		if oid, err := bson.ObjectIDFromHex(id); err == nil {
			oids = append(oids, oid)
		}
	}

	if len(oids) == 0 {
		return map[string]int{}, nil
	}

	cursor, err := r.voteColl.Find(ctx, bson.M{
		"commentId": bson.M{"$in": oids},
		"userId":    userOID,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	votes := make(map[string]int)
	for cursor.Next(ctx) {
		var vote models.CampaignCommentVote
		if err := cursor.Decode(&vote); err != nil {
			continue
		}
		votes[vote.CommentID.Hex()] = vote.Value
	}

	return votes, nil
}

// CountByCampaign returns the total comment count for a campaign.
func (r *CampaignCommentRepository) CountByCampaign(ctx context.Context, campaignID string) (int64, error) {
	campaignOID, err := bson.ObjectIDFromHex(campaignID)
	if err != nil {
		return 0, nil
	}
	return r.coll.CountDocuments(ctx, bson.M{"campaignId": campaignOID})
}

func (r *CampaignCommentRepository) getSortField(sort string) bson.D {
	switch sort {
	case "votes":
		return bson.D{{Key: "votes", Value: -1}, {Key: "createdAt", Value: -1}}
	case "newest":
		fallthrough
	default:
		return bson.D{{Key: "createdAt", Value: -1}}
	}
}
