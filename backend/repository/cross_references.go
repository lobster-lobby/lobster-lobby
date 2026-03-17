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

// ErrDuplicateCrossReference is returned when a cross-reference already exists.
var ErrDuplicateCrossReference = errors.New("cross-reference already exists")

type CrossReferenceRepository struct {
	refs     *mongo.Collection
	policies *mongo.Collection
	research *mongo.Collection
	debates  *mongo.Collection
}

func NewCrossReferenceRepository(db *MongoDB) *CrossReferenceRepository {
	return &CrossReferenceRepository{
		refs:     db.Database.Collection("cross_references"),
		policies: db.Database.Collection("policies"),
		research: db.Database.Collection("research"),
		debates:  db.Database.Collection("debates"),
	}
}

func (r *CrossReferenceRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.refs.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "sourceType", Value: 1}, {Key: "sourceId", Value: 1}, {Key: "targetType", Value: 1}, {Key: "targetId", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{Keys: bson.D{{Key: "sourceType", Value: 1}, {Key: "sourceId", Value: 1}}},
		{Keys: bson.D{{Key: "targetType", Value: 1}, {Key: "targetId", Value: 1}}},
	})
	return err
}

func (r *CrossReferenceRepository) Create(ctx context.Context, ref *models.CrossReference) (*models.CrossReferenceResponse, error) {
	now := time.Now().UTC()
	ref.CreatedAt = now
	if ref.ID.IsZero() {
		ref.ID = bson.NewObjectID()
	}

	// Check reverse direction (B→A when A→B exists)
	reverseCount, err := r.refs.CountDocuments(ctx, bson.M{
		"sourceType": ref.TargetType,
		"sourceId":   ref.TargetID,
		"targetType": ref.SourceType,
		"targetId":   ref.SourceID,
	})
	if err != nil {
		return nil, err
	}
	if reverseCount > 0 {
		return nil, ErrDuplicateCrossReference
	}

	_, err = r.refs.InsertOne(ctx, ref)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrDuplicateCrossReference
		}
		return nil, err
	}

	return r.toResponse(ctx, ref)
}

// GetForEntity returns all cross-references where the entity appears as source OR target (bidirectional).
func (r *CrossReferenceRepository) GetForEntity(ctx context.Context, entityType string, entityID bson.ObjectID) ([]*models.CrossReferenceResponse, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"sourceType": entityType, "sourceId": entityID},
			{"targetType": entityType, "targetId": entityID},
		},
	}

	findOpts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := r.refs.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []models.CrossReference
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}

	responses := make([]*models.CrossReferenceResponse, len(items))
	for i := range items {
		resp, _ := r.toResponse(ctx, &items[i])
		responses[i] = resp
	}
	return responses, nil
}

func (r *CrossReferenceRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	result, err := r.refs.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *CrossReferenceRepository) GetByID(ctx context.Context, id bson.ObjectID) (*models.CrossReference, error) {
	var ref models.CrossReference
	err := r.refs.FindOne(ctx, bson.M{"_id": id}).Decode(&ref)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ref, nil
}

func (r *CrossReferenceRepository) resolveTitle(ctx context.Context, entityType string, entityID bson.ObjectID) string {
	switch entityType {
	case "policy":
		var policy models.Policy
		if err := r.policies.FindOne(ctx, bson.M{"_id": entityID}).Decode(&policy); err == nil {
			return policy.Title
		}
	case "research":
		var research models.Research
		if err := r.research.FindOne(ctx, bson.M{"_id": entityID}).Decode(&research); err == nil {
			return research.Title
		}
	case "debate":
		var debate models.Debate
		if err := r.debates.FindOne(ctx, bson.M{"_id": entityID}).Decode(&debate); err == nil {
			return debate.Title
		}
	}
	return ""
}

func (r *CrossReferenceRepository) toResponse(ctx context.Context, ref *models.CrossReference) (*models.CrossReferenceResponse, error) {
	resp := &models.CrossReferenceResponse{
		ID:         ref.ID,
		SourceType: ref.SourceType,
		SourceID:   ref.SourceID,
		TargetType: ref.TargetType,
		TargetID:   ref.TargetID,
		CreatedBy:  ref.CreatedBy,
		CreatedAt:  ref.CreatedAt,
	}

	resp.SourceTitle = r.resolveTitle(ctx, ref.SourceType, ref.SourceID)
	resp.TargetTitle = r.resolveTitle(ctx, ref.TargetType, ref.TargetID)

	return resp, nil
}
