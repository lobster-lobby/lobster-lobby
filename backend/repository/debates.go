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

type DebateRepository struct {
	debates   *mongo.Collection
	arguments *mongo.Collection
	votes     *mongo.Collection
	users     *mongo.Collection
}

func NewDebateRepository(db *MongoDB) *DebateRepository {
	return &DebateRepository{
		debates:   db.Database.Collection("debates"),
		arguments: db.Database.Collection("debate_arguments"),
		votes:     db.Database.Collection("debate_votes"),
		users:     db.Database.Collection("users"),
	}
}

func (r *DebateRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.debates.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "slug", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{Keys: bson.D{{Key: "createdAt", Value: -1}}},
	})
	if err != nil {
		return err
	}

	_, err = r.arguments.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "debateId", Value: 1}, {Key: "score", Value: -1}}},
		{Keys: bson.D{{Key: "debateId", Value: 1}, {Key: "createdAt", Value: -1}}},
	})
	if err != nil {
		return err
	}

	_, err = r.votes.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "userId", Value: 1}, {Key: "argumentId", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{Keys: bson.D{{Key: "argumentId", Value: 1}}},
	})
	return err
}

// CreateDebate creates a new debate.
func (r *DebateRepository) CreateDebate(ctx context.Context, debate *models.Debate) (*models.Debate, error) {
	now := time.Now().UTC()
	debate.CreatedAt = now
	debate.UpdatedAt = now
	if debate.ID.IsZero() {
		debate.ID = bson.NewObjectID()
	}
	if debate.Status == "" {
		debate.Status = "open"
	}

	_, err := r.debates.InsertOne(ctx, debate)
	if err != nil {
		return nil, err
	}
	return debate, nil
}

// GetDebateBySlug retrieves a debate by its slug.
func (r *DebateRepository) GetDebateBySlug(ctx context.Context, slug string) (*models.DebateResponse, error) {
	var debate models.Debate
	err := r.debates.FindOne(ctx, bson.M{"slug": slug}).Decode(&debate)
	if err != nil {
		return nil, err
	}

	resp := &models.DebateResponse{Debate: debate}
	r.enrichDebate(ctx, resp)
	return resp, nil
}

// ListDebates returns paginated debates.
func (r *DebateRepository) ListDebates(ctx context.Context, page, perPage int) ([]models.DebateResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	total, err := r.debates.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(int64((page - 1) * perPage)).
		SetLimit(int64(perPage))

	cursor, err := r.debates.Find(ctx, bson.M{}, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var debates []models.Debate
	if err := cursor.All(ctx, &debates); err != nil {
		return nil, 0, err
	}

	responses := make([]models.DebateResponse, len(debates))
	for i, d := range debates {
		responses[i] = models.DebateResponse{Debate: d}
		r.enrichDebate(ctx, &responses[i])
	}

	return responses, total, nil
}

// CreateArgument creates a new argument on a debate.
func (r *DebateRepository) CreateArgument(ctx context.Context, arg *models.Argument) (*models.ArgumentResponse, error) {
	now := time.Now().UTC()
	arg.CreatedAt = now
	arg.UpdatedAt = now
	if arg.ID.IsZero() {
		arg.ID = bson.NewObjectID()
	}

	_, err := r.arguments.InsertOne(ctx, arg)
	if err != nil {
		return nil, err
	}

	resp := &models.ArgumentResponse{Argument: *arg}
	r.enrichArgument(ctx, resp)
	return resp, nil
}

// ListArguments returns arguments for a debate, sorted by the specified method.
func (r *DebateRepository) ListArguments(ctx context.Context, debateID bson.ObjectID, sort string) ([]models.ArgumentResponse, error) {
	sortDoc := r.getArgumentSortDoc(sort)

	cursor, err := r.arguments.Find(ctx,
		bson.M{"debateId": debateID},
		options.Find().SetSort(sortDoc),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var args []models.Argument
	if err := cursor.All(ctx, &args); err != nil {
		return nil, err
	}

	responses := make([]models.ArgumentResponse, len(args))
	for i, a := range args {
		responses[i] = models.ArgumentResponse{Argument: a}
		r.enrichArgument(ctx, &responses[i])
	}

	return responses, nil
}

// ToggleVote handles vote toggle: same vote removes it, different vote changes it.
func (r *DebateRepository) ToggleVote(ctx context.Context, userID, argumentID, debateID bson.ObjectID, value int) (newValue int, err error) {
	// Get existing vote
	var existing models.Vote
	existErr := r.votes.FindOne(ctx, bson.M{
		"userId":     userID,
		"argumentId": argumentID,
	}).Decode(&existing)

	oldValue := 0
	if existErr == nil {
		oldValue = existing.Value
	}

	// Compute new value using toggle logic
	newValue, delta := ComputeVoteToggle(oldValue, value)

	if newValue == 0 {
		// Remove vote
		_, _ = r.votes.DeleteOne(ctx, bson.M{
			"userId":     userID,
			"argumentId": argumentID,
		})
	} else {
		// Upsert vote
		_, err = r.votes.UpdateOne(ctx,
			bson.M{"userId": userID, "argumentId": argumentID},
			bson.M{"$set": bson.M{
				"userId":     userID,
				"argumentId": argumentID,
				"debateId":   debateID,
				"value":      newValue,
				"createdAt":  time.Now().UTC(),
			}},
			options.UpdateOne().SetUpsert(true),
		)
		if err != nil {
			return 0, err
		}
	}

	// Update argument vote counters
	inc := bson.M{}
	if oldValue == 1 {
		inc["upvotes"] = -1
	} else if oldValue == -1 {
		inc["downvotes"] = -1
	}
	if newValue == 1 {
		inc["upvotes"] = getOrDefault(inc, "upvotes") + 1
	} else if newValue == -1 {
		inc["downvotes"] = getOrDefault(inc, "downvotes") + 1
	}

	if len(inc) > 0 {
		_, err = r.arguments.UpdateOne(ctx,
			bson.M{"_id": argumentID},
			bson.M{"$inc": inc},
		)
		if err != nil {
			return 0, err
		}
	}

	// Update net score
	_ = delta // delta used for consistency check, score recomputed from DB
	var arg models.Argument
	if err := r.arguments.FindOne(ctx, bson.M{"_id": argumentID}).Decode(&arg); err == nil {
		score := arg.Upvotes - arg.Downvotes
		_, _ = r.arguments.UpdateOne(ctx,
			bson.M{"_id": argumentID},
			bson.M{"$set": bson.M{"score": score}},
		)
	}

	return newValue, nil
}

// GetUserVote returns the user's current vote on an argument (0 if none).
func (r *DebateRepository) GetUserVote(ctx context.Context, userID, argumentID bson.ObjectID) int {
	var vote models.Vote
	err := r.votes.FindOne(ctx, bson.M{
		"userId":     userID,
		"argumentId": argumentID,
	}).Decode(&vote)
	if err != nil {
		return 0
	}
	return vote.Value
}

// ComputeVoteToggle returns the new vote value and delta for score updates.
// Exported for testing.
func ComputeVoteToggle(existingVote, requestedVote int) (newVote, delta int) {
	if existingVote == 0 {
		return requestedVote, requestedVote
	}
	if existingVote == requestedVote {
		return 0, -existingVote
	}
	return requestedVote, requestedVote - existingVote
}

func (r *DebateRepository) getArgumentSortDoc(sort string) bson.D {
	switch sort {
	case "newest":
		return bson.D{{Key: "createdAt", Value: -1}}
	case "top":
		return bson.D{{Key: "score", Value: -1}, {Key: "createdAt", Value: -1}}
	case "controversial":
		// Controversial: high total votes but close to even (score near 0)
		// We sort by total votes desc, then by absolute score asc (closer to 0 = more controversial)
		// MongoDB doesn't support computed sorts easily, so we use a pipeline
		// Fallback: sort by downvotes desc (most contested), then upvotes desc
		return bson.D{{Key: "downvotes", Value: -1}, {Key: "upvotes", Value: -1}}
	default:
		return bson.D{{Key: "score", Value: -1}, {Key: "createdAt", Value: -1}}
	}
}

// ListArgumentsControversial returns arguments sorted by controversy score.
// Controversy = high engagement + divided votes. Uses aggregation pipeline.
func (r *DebateRepository) ListArgumentsControversial(ctx context.Context, debateID bson.ObjectID) ([]models.ArgumentResponse, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{"debateId": debateID}},
		bson.M{"$addFields": bson.M{
			"totalVotes": bson.M{"$add": bson.A{"$upvotes", "$downvotes"}},
			"absScore": bson.M{"$abs": bson.M{"$subtract": bson.A{"$upvotes", "$downvotes"}}},
		}},
		// Controversial score: high total votes, low absolute score
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
	}

	cursor, err := r.arguments.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var args []models.Argument
	if err := cursor.All(ctx, &args); err != nil {
		return nil, err
	}

	responses := make([]models.ArgumentResponse, len(args))
	for i, a := range args {
		responses[i] = models.ArgumentResponse{Argument: a}
		r.enrichArgument(ctx, &responses[i])
	}

	return responses, nil
}

func (r *DebateRepository) enrichDebate(ctx context.Context, dr *models.DebateResponse) {
	var user models.User
	if err := r.users.FindOne(ctx, bson.M{"_id": dr.CreatorID}).Decode(&user); err == nil {
		dr.CreatorUsername = user.Username
	}

	count, err := r.arguments.CountDocuments(ctx, bson.M{"debateId": dr.ID})
	if err == nil {
		dr.ArgumentCount = int(count)
	}
}

func (r *DebateRepository) enrichArgument(ctx context.Context, ar *models.ArgumentResponse) {
	var user models.User
	if err := r.users.FindOne(ctx, bson.M{"_id": ar.AuthorID}).Decode(&user); err == nil {
		ar.AuthorUsername = user.Username
		ar.AuthorRepTier = user.Reputation.Tier
	}
}

// ControversyScore computes a controversy score for sorting.
// Higher values = more controversial (high engagement, close to even).
func ControversyScore(upvotes, downvotes int) float64 {
	total := float64(upvotes + downvotes)
	if total == 0 {
		return 0
	}
	absScore := math.Abs(float64(upvotes - downvotes))
	return total / (1 + absScore)
}
