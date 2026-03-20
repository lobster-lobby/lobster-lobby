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

type DraftRepository struct {
	drafts       *mongo.Collection
	endorsements *mongo.Collection
	comments     *mongo.Collection
}

func NewDraftRepository(db *MongoDB) *DraftRepository {
	return &DraftRepository{
		drafts:       db.Database.Collection("drafts"),
		endorsements: db.Database.Collection("draft_endorsements"),
		comments:     db.Database.Collection("draft_comments"),
	}
}

func (r *DraftRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.drafts.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "policyId", Value: 1}, {Key: "endorsements", Value: -1}}},
		{Keys: bson.D{{Key: "policyId", Value: 1}, {Key: "createdAt", Value: -1}}},
	})
	if err != nil {
		return err
	}
	_, err = r.endorsements.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "draftId", Value: 1}, {Key: "userId", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		return err
	}
	_, err = r.comments.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "draftId", Value: 1}, {Key: "createdAt", Value: 1}}},
	})
	return err
}

func (r *DraftRepository) Create(ctx context.Context, draft *models.Draft) (*models.Draft, error) {
	now := time.Now().UTC()
	draft.CreatedAt = now
	draft.UpdatedAt = now
	if draft.ID.IsZero() {
		draft.ID = bson.NewObjectID()
	}
	if draft.Status == "" {
		draft.Status = "draft"
	}
	if draft.Version == 0 {
		draft.Version = 1
	}
	_, err := r.drafts.InsertOne(ctx, draft)
	if err != nil {
		return nil, err
	}
	return draft, nil
}

func (r *DraftRepository) ListByPolicy(ctx context.Context, policyID bson.ObjectID, sort string) ([]models.Draft, error) {
	var sortDoc bson.D
	if sort == "newest" {
		sortDoc = bson.D{{Key: "createdAt", Value: -1}}
	} else {
		sortDoc = bson.D{{Key: "endorsements", Value: -1}, {Key: "createdAt", Value: -1}}
	}

	filter := bson.M{"policyId": policyID, "status": bson.M{"$ne": "archived"}}
	cursor, err := r.drafts.Find(ctx, filter, options.Find().SetSort(sortDoc))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var drafts []models.Draft
	if err := cursor.All(ctx, &drafts); err != nil {
		return nil, err
	}
	if drafts == nil {
		drafts = []models.Draft{}
	}
	return drafts, nil
}

func (r *DraftRepository) GetByID(ctx context.Context, id bson.ObjectID) (*models.Draft, error) {
	var draft models.Draft
	err := r.drafts.FindOne(ctx, bson.M{"_id": id}).Decode(&draft)
	if err != nil {
		return nil, err
	}
	return &draft, nil
}

func (r *DraftRepository) Update(ctx context.Context, id bson.ObjectID, title, content, category, status string) (*models.Draft, error) {
	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"updatedAt": now,
		},
		"$inc": bson.M{"version": 1},
	}
	set := update["$set"].(bson.M)
	if title != "" {
		set["title"] = title
	}
	if content != "" {
		set["content"] = content
	}
	if category != "" {
		set["category"] = category
	}
	if status != "" {
		set["status"] = status
	}

	var updated models.Draft
	err := r.drafts.FindOneAndUpdate(ctx,
		bson.M{"_id": id},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *DraftRepository) Archive(ctx context.Context, id bson.ObjectID) error {
	_, err := r.drafts.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": "archived", "updatedAt": time.Now().UTC()}},
	)
	return err
}

// ToggleEndorsement adds or removes an endorsement. Returns whether endorsed after the toggle.
func (r *DraftRepository) ToggleEndorsement(ctx context.Context, draftID, userID bson.ObjectID) (bool, error) {
	var existing models.DraftEndorsement
	err := r.endorsements.FindOne(ctx, bson.M{"draftId": draftID, "userId": userID}).Decode(&existing)

	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return false, err
	}

	if err == nil {
		// Already endorsed — remove it
		_, err = r.endorsements.DeleteOne(ctx, bson.M{"draftId": draftID, "userId": userID})
		if err != nil {
			return false, err
		}
		_, err = r.drafts.UpdateOne(ctx,
			bson.M{"_id": draftID},
			bson.M{"$inc": bson.M{"endorsements": -1}, "$set": bson.M{"updatedAt": time.Now().UTC()}},
		)
		return false, err
	}

	// Not endorsed — add it
	endorsement := models.DraftEndorsement{
		ID:        bson.NewObjectID(),
		DraftID:   draftID,
		UserID:    userID,
		CreatedAt: time.Now().UTC(),
	}
	_, err = r.endorsements.InsertOne(ctx, endorsement)
	if err != nil {
		return false, err
	}
	_, err = r.drafts.UpdateOne(ctx,
		bson.M{"_id": draftID},
		bson.M{"$inc": bson.M{"endorsements": 1}, "$set": bson.M{"updatedAt": time.Now().UTC()}},
	)
	return true, err
}

func (r *DraftRepository) IsEndorsedBy(ctx context.Context, draftID, userID bson.ObjectID) bool {
	count, err := r.endorsements.CountDocuments(ctx, bson.M{"draftId": draftID, "userId": userID})
	return err == nil && count > 0
}

func (r *DraftRepository) ListComments(ctx context.Context, draftID bson.ObjectID) ([]models.DraftComment, error) {
	cursor, err := r.comments.Find(ctx,
		bson.M{"draftId": draftID},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comments []models.DraftComment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}
	if comments == nil {
		comments = []models.DraftComment{}
	}
	return comments, nil
}

func (r *DraftRepository) AddComment(ctx context.Context, comment *models.DraftComment) (*models.DraftComment, error) {
	if comment.ID.IsZero() {
		comment.ID = bson.NewObjectID()
	}
	comment.CreatedAt = time.Now().UTC()
	_, err := r.comments.InsertOne(ctx, comment)
	if err != nil {
		return nil, err
	}
	return comment, nil
}
