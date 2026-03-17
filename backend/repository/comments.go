package repository

import (
	"context"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/lobster-lobby/lobster-lobby/models"
)

type CommentRepository struct {
	comments  *mongo.Collection
	reactions *mongo.Collection
	stances   *mongo.Collection
	users     *mongo.Collection
}

func NewCommentRepository(db *MongoDB) *CommentRepository {
	return &CommentRepository{
		comments:  db.Database.Collection("comments"),
		reactions: db.Database.Collection("reactions"),
		stances:   db.Database.Collection("stances"),
		users:     db.Database.Collection("users"),
	}
}

func (r *CommentRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.comments.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "policyId", Value: 1}, {Key: "parentId", Value: 1}, {Key: "score", Value: -1}}},
		{Keys: bson.D{{Key: "policyId", Value: 1}, {Key: "parentId", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "authorId", Value: 1}}},
	})
	if err != nil {
		return err
	}

	_, err = r.reactions.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "userId", Value: 1}, {Key: "entityId", Value: 1}, {Key: "entityType", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		return err
	}

	_, err = r.stances.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "userId", Value: 1}, {Key: "policyId", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	return err
}

func (r *CommentRepository) Create(ctx context.Context, comment *models.Comment) (*models.Comment, error) {
	now := time.Now().UTC()
	comment.CreatedAt = now
	comment.UpdatedAt = now
	if comment.ID.IsZero() {
		comment.ID = bson.NewObjectID()
	}

	_, err := r.comments.InsertOne(ctx, comment)
	if err != nil {
		return nil, err
	}

	// Increment parent's reply count
	if comment.ParentID != nil {
		_, _ = r.comments.UpdateOne(ctx,
			bson.M{"_id": *comment.ParentID},
			bson.M{"$inc": bson.M{"replyCount": 1}},
		)
	}

	return comment, nil
}

func (r *CommentRepository) FindByPolicy(ctx context.Context, policyID bson.ObjectID, sort, position string, page, perPage int) ([]models.CommentResponse, int64, map[string]int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	filter := bson.M{"policyId": policyID, "parentId": nil}
	if position != "" && position != "all" {
		filter["position"] = position
	}

	total, err := r.comments.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, nil, err
	}

	sortDoc := r.getSortDoc(sort)
	findOpts := options.Find().
		SetSort(sortDoc).
		SetSkip(int64((page - 1) * perPage)).
		SetLimit(int64(perPage))

	cursor, err := r.comments.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, nil, err
	}
	defer cursor.Close(ctx)

	var comments []models.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, 0, nil, err
	}

	responses := make([]models.CommentResponse, len(comments))
	for i, c := range comments {
		responses[i] = models.CommentResponse{Comment: c}
		r.enrichComment(ctx, &responses[i])
	}

	// Position counts (for entire policy, not filtered)
	positions := r.countPositions(ctx, policyID)

	return responses, total, positions, nil
}

func (r *CommentRepository) FindByID(ctx context.Context, id bson.ObjectID) (*models.Comment, error) {
	var c models.Comment
	err := r.comments.FindOne(ctx, bson.M{"_id": id}).Decode(&c)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &c, err
}

func (r *CommentRepository) FindReplies(ctx context.Context, parentID bson.ObjectID) ([]models.CommentResponse, error) {
	cursor, err := r.comments.Find(ctx,
		bson.M{"parentId": parentID},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comments []models.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}

	responses := make([]models.CommentResponse, len(comments))
	for i, c := range comments {
		responses[i] = models.CommentResponse{Comment: c}
		r.enrichComment(ctx, &responses[i])
	}

	return responses, nil
}

func (r *CommentRepository) Update(ctx context.Context, id, authorID bson.ObjectID, content string) error {
	now := time.Now().UTC()
	result, err := r.comments.UpdateOne(ctx,
		bson.M{"_id": id, "authorId": authorID},
		bson.M{"$set": bson.M{"content": content, "editedAt": now, "updatedAt": now}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *CommentRepository) React(ctx context.Context, userID, commentID bson.ObjectID, value int) error {
	// Get existing reaction
	var existing models.Reaction
	err := r.reactions.FindOne(ctx, bson.M{
		"userId":     userID,
		"entityId":   commentID,
		"entityType": "comment",
	}).Decode(&existing)

	oldValue := 0
	if err == nil {
		oldValue = existing.Value
	}

	if value == 0 {
		// Remove reaction
		_, _ = r.reactions.DeleteOne(ctx, bson.M{
			"userId":     userID,
			"entityId":   commentID,
			"entityType": "comment",
		})
	} else {
		// Upsert reaction
		_, err = r.reactions.UpdateOne(ctx,
			bson.M{"userId": userID, "entityId": commentID, "entityType": "comment"},
			bson.M{"$set": bson.M{
				"userId":     userID,
				"entityId":   commentID,
				"entityType": "comment",
				"value":      value,
				"createdAt":  time.Now().UTC(),
			}},
			options.UpdateOne().SetUpsert(true),
		)
		if err != nil {
			return err
		}
	}

	// Update comment counters
	inc := bson.M{}
	if oldValue == 1 {
		inc["upvotes"] = -1
	} else if oldValue == -1 {
		inc["downvotes"] = -1
	}
	if value == 1 {
		inc["upvotes"] = getOrDefault(inc, "upvotes") + 1
	} else if value == -1 {
		inc["downvotes"] = getOrDefault(inc, "downvotes") + 1
	}

	if len(inc) > 0 {
		_, err = r.comments.UpdateOne(ctx,
			bson.M{"_id": commentID},
			bson.M{"$inc": inc},
		)
		if err != nil {
			return err
		}
	}

	// Recalculate score and Wilson score
	var comment models.Comment
	if err := r.comments.FindOne(ctx, bson.M{"_id": commentID}).Decode(&comment); err == nil {
		score := comment.Upvotes - comment.Downvotes
		wilson := WilsonScore(comment.Upvotes, comment.Downvotes, comment.CreatedAt)
		_, _ = r.comments.UpdateOne(ctx,
			bson.M{"_id": commentID},
			bson.M{"$set": bson.M{"score": score, "wilsonScore": wilson}},
		)
	}

	return nil
}

func getOrDefault(m bson.M, key string, ) int {
	if v, ok := m[key]; ok {
		return v.(int)
	}
	return 0
}

func (r *CommentRepository) SetStance(ctx context.Context, userID, policyID bson.ObjectID, position string) error {
	_, err := r.stances.UpdateOne(ctx,
		bson.M{"userId": userID, "policyId": policyID},
		bson.M{"$set": bson.M{
			"userId":    userID,
			"policyId":  policyID,
			"position":  position,
			"updatedAt": time.Now().UTC(),
		}},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (r *CommentRepository) GetStance(ctx context.Context, userID, policyID bson.ObjectID) (*models.Stance, error) {
	var s models.Stance
	err := r.stances.FindOne(ctx, bson.M{"userId": userID, "policyId": policyID}).Decode(&s)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &s, err
}

func (r *CommentRepository) GetUserReaction(ctx context.Context, userID, entityID bson.ObjectID) int {
	var reaction models.Reaction
	err := r.reactions.FindOne(ctx, bson.M{
		"userId":     userID,
		"entityId":   entityID,
		"entityType": "comment",
	}).Decode(&reaction)
	if err != nil {
		return 0
	}
	return reaction.Value
}

func (r *CommentRepository) enrichComment(ctx context.Context, cr *models.CommentResponse) {
	var user models.User
	if err := r.users.FindOne(ctx, bson.M{"_id": cr.AuthorID}).Decode(&user); err == nil {
		cr.AuthorUsername = user.Username
		cr.AuthorRepTier = user.Reputation.Tier
	}
}

func (r *CommentRepository) countPositions(ctx context.Context, policyID bson.ObjectID) map[string]int {
	positions := map[string]int{"support": 0, "oppose": 0, "neutral": 0}

	pipeline := bson.A{
		bson.M{"$match": bson.M{"policyId": policyID, "parentId": nil}},
		bson.M{"$group": bson.M{"_id": "$position", "count": bson.M{"$sum": 1}}},
	}

	cursor, err := r.comments.Aggregate(ctx, pipeline)
	if err != nil {
		return positions
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if cursor.Decode(&result) == nil {
			positions[result.ID] = result.Count
		}
	}
	return positions
}

func (r *CommentRepository) getSortDoc(sort string) bson.D {
	switch sort {
	case "newest":
		return bson.D{{Key: "createdAt", Value: -1}}
	case "top":
		return bson.D{{Key: "score", Value: -1}, {Key: "createdAt", Value: -1}}
	case "discussed":
		return bson.D{{Key: "replyCount", Value: -1}, {Key: "createdAt", Value: -1}}
	case "best":
		return bson.D{{Key: "wilsonScore", Value: -1}, {Key: "createdAt", Value: -1}}
	default:
		return bson.D{{Key: "score", Value: -1}, {Key: "createdAt", Value: -1}}
	}
}

// FindByPolicyControversial returns top-level comments sorted by controversy score.
// Controversial = high engagement + close ratio: totalVotes / (1 + |upvotes - downvotes|).
func (r *CommentRepository) FindByPolicyControversial(ctx context.Context, policyID bson.ObjectID, position string, page, perPage int) ([]models.CommentResponse, int64, map[string]int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	filter := bson.M{"policyId": policyID, "parentId": nil}
	if position != "" && position != "all" {
		filter["position"] = position
	}

	total, err := r.comments.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, nil, err
	}

	pipeline := bson.A{
		bson.M{"$match": filter},
		bson.M{"$addFields": bson.M{
			"totalVotes": bson.M{"$add": bson.A{"$upvotes", "$downvotes"}},
			"absScore":   bson.M{"$abs": bson.M{"$subtract": bson.A{"$upvotes", "$downvotes"}}},
		}},
		bson.M{"$addFields": bson.M{
			"controversyScore": bson.M{
				"$cond": bson.M{
					"if":   bson.M{"$eq": bson.A{"$totalVotes", 0}},
					"then": 0,
					"else": bson.M{
						"$divide": bson.A{
							"$totalVotes",
							bson.M{"$add": bson.A{1, "$absScore"}},
						},
					},
				},
			},
		}},
		bson.M{"$sort": bson.D{{Key: "controversyScore", Value: -1}, {Key: "totalVotes", Value: -1}}},
		bson.M{"$skip": int64((page - 1) * perPage)},
		bson.M{"$limit": int64(perPage)},
	}

	cursor, err := r.comments.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, nil, err
	}
	defer cursor.Close(ctx)

	var comments []models.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, 0, nil, err
	}

	responses := make([]models.CommentResponse, len(comments))
	for i, c := range comments {
		responses[i] = models.CommentResponse{Comment: c}
		r.enrichComment(ctx, &responses[i])
	}

	positions := r.countPositions(ctx, policyID)

	return responses, total, positions, nil
}

// WilsonScore calculates the lower bound of Wilson score confidence interval.
// Used for "best" comment sorting with recency decay.
func WilsonScore(upvotes, downvotes int, createdAt time.Time) float64 {
	n := float64(upvotes + downvotes)
	if n == 0 {
		return 0
	}

	z := 1.96 // 95% confidence
	p := float64(upvotes) / n
	denominator := 1 + z*z/n
	centre := p + z*z/(2*n)
	spread := z * math.Sqrt((p*(1-p)+z*z/(4*n))/n)

	wilson := (centre - spread) / denominator

	// Recency decay: half-life of 7 days
	ageHours := time.Since(createdAt).Hours()
	decay := math.Exp(-0.004 * ageHours) // ~0.5 at 7 days

	return wilson * decay
}
