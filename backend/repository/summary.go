package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/lobster-lobby/lobster-lobby/models"
)

type SummaryPointRepository struct {
	points *mongo.Collection
	users  *mongo.Collection
}

func NewSummaryPointRepository(db *MongoDB) *SummaryPointRepository {
	return &SummaryPointRepository{
		points: db.Database.Collection("summaryPoints"),
		users:  db.Database.Collection("users"),
	}
}

func (r *SummaryPointRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.points.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "policyId", Value: 1}, {Key: "visible", Value: 1}, {Key: "bridgingScore", Value: -1}}},
		{Keys: bson.D{{Key: "policyId", Value: 1}, {Key: "authorId", Value: 1}}},
	})
	return err
}

func (r *SummaryPointRepository) Create(ctx context.Context, point *models.SummaryPoint) (*models.SummaryPoint, error) {
	now := time.Now().UTC()
	point.CreatedAt = now
	point.UpdatedAt = now
	if point.ID.IsZero() {
		point.ID = bson.NewObjectID()
	}
	if point.Endorsements == nil {
		point.Endorsements = []models.Endorsement{}
	}
	point.BridgingScore = 0
	point.Visible = false

	_, err := r.points.InsertOne(ctx, point)
	if err != nil {
		return nil, err
	}
	return point, nil
}

func (r *SummaryPointRepository) FindByID(ctx context.Context, id bson.ObjectID) (*models.SummaryPoint, error) {
	var p models.SummaryPoint
	err := r.points.FindOne(ctx, bson.M{"_id": id}).Decode(&p)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &p, err
}

func (r *SummaryPointRepository) ListByPolicy(ctx context.Context, policyID bson.ObjectID, includeHidden bool) ([]models.SummaryPoint, error) {
	filter := bson.M{"policyId": policyID}
	if !includeHidden {
		filter["visible"] = true
	}

	cursor, err := r.points.Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "bridgingScore", Value: -1}, {Key: "createdAt", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var points []models.SummaryPoint
	if err := cursor.All(ctx, &points); err != nil {
		return nil, err
	}
	return points, nil
}

func (r *SummaryPointRepository) AddEndorsement(ctx context.Context, pointID bson.ObjectID, endorsement models.Endorsement) error {
	// Remove any existing endorsement from this user first
	_, _ = r.points.UpdateOne(ctx,
		bson.M{"_id": pointID},
		bson.M{"$pull": bson.M{"endorsements": bson.M{"userId": endorsement.UserID}}},
	)

	// Add the new endorsement
	_, err := r.points.UpdateOne(ctx,
		bson.M{"_id": pointID},
		bson.M{
			"$push": bson.M{"endorsements": endorsement},
			"$set":  bson.M{"updatedAt": time.Now().UTC()},
		},
	)
	if err != nil {
		return err
	}

	return r.RecalculateScore(ctx, pointID)
}

func (r *SummaryPointRepository) RemoveEndorsement(ctx context.Context, pointID, userID bson.ObjectID) error {
	result, err := r.points.UpdateOne(ctx,
		bson.M{"_id": pointID},
		bson.M{
			"$pull": bson.M{"endorsements": bson.M{"userId": userID}},
			"$set":  bson.M{"updatedAt": time.Now().UTC()},
		},
	)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return r.RecalculateScore(ctx, pointID)
}

func (r *SummaryPointRepository) RecalculateScore(ctx context.Context, pointID bson.ObjectID) error {
	point, err := r.FindByID(ctx, pointID)
	if err != nil || point == nil {
		return err
	}

	score := models.CalculateBridgingScore(point.Position, point.Endorsements)
	visible := score >= models.BridgingVisibilityThreshold

	_, err = r.points.UpdateOne(ctx,
		bson.M{"_id": pointID},
		bson.M{"$set": bson.M{"bridgingScore": score, "visible": visible}},
	)
	return err
}

func (r *SummaryPointRepository) EnrichResponse(ctx context.Context, point models.SummaryPoint, userID *bson.ObjectID) models.SummaryPointResponse {
	resp := models.SummaryPointResponse{
		SummaryPoint: point,
		EndorseCount: len(point.Endorsements),
	}

	// Count cross-position endorsements
	for _, e := range point.Endorsements {
		if e.Position != point.Position && e.Position != models.PositionConsensus {
			resp.CrossCount++
		}
	}

	// Check if current user endorsed
	if userID != nil {
		for _, e := range point.Endorsements {
			if e.UserID == *userID {
				resp.UserEndorsed = true
				break
			}
		}
	}

	// Fetch author info
	var user models.User
	if err := r.users.FindOne(ctx, bson.M{"_id": point.AuthorID}).Decode(&user); err == nil {
		resp.AuthorUsername = user.Username
		resp.AuthorRepTier = user.Reputation.Tier
	}

	return resp
}
