package repository

import (
	"context"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/lobster-lobby/lobster-lobby/models"
)

type ResearchRepository struct {
	research  *mongo.Collection
	reactions *mongo.Collection
	policies  *mongo.Collection
	users     *mongo.Collection
}

func NewResearchRepository(db *MongoDB) *ResearchRepository {
	return &ResearchRepository{
		research:  db.Database.Collection("research"),
		reactions: db.Database.Collection("reactions"),
		policies:  db.Database.Collection("policies"),
		users:     db.Database.Collection("users"),
	}
}

func (r *ResearchRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.research.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "policyId", Value: 1}, {Key: "score", Value: -1}}},
		{Keys: bson.D{{Key: "policyId", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "policyId", Value: 1}, {Key: "citedBy", Value: -1}}},
		{Keys: bson.D{{Key: "authorId", Value: 1}}},
	})
	return err
}

// IsInstitutional checks if a URL is from an institutional source
func (r *ResearchRepository) IsInstitutional(urlStr string) bool {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	host := strings.ToLower(parsed.Hostname())

	// Check TLD
	if strings.HasSuffix(host, ".gov") || strings.HasSuffix(host, ".edu") {
		return true
	}

	// Check known institutional domains
	institutionalDomains := []string{
		"reuters.com",
		"apnews.com",
		"bbc.com",
		"nature.com",
		"sciencedirect.com",
		"pubmed.ncbi.nlm.nih.gov",
		"cbo.gov",
		"gao.gov",
		"who.int",
		"un.org",
		"worldbank.org",
		"imf.org",
		"brookings.edu",
		"rand.org",
		"pewresearch.org",
	}

	for _, domain := range institutionalDomains {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}

	return false
}

func (r *ResearchRepository) Create(ctx context.Context, research *models.Research) (*models.ResearchResponse, error) {
	now := time.Now().UTC()
	research.CreatedAt = now
	research.UpdatedAt = now
	if research.ID.IsZero() {
		research.ID = bson.NewObjectID()
	}

	// Detect institutional status for each source
	for i := range research.Sources {
		research.Sources[i].Institutional = r.IsInstitutional(research.Sources[i].URL)
	}

	_, err := r.research.InsertOne(ctx, research)
	if err != nil {
		return nil, err
	}

	// Increment policy engagement.researchCount
	_, _ = r.policies.UpdateOne(ctx,
		bson.M{"_id": research.PolicyID},
		bson.M{"$inc": bson.M{"engagement.researchCount": 1}},
	)

	return r.toResponse(ctx, research, nil)
}

func (r *ResearchRepository) List(ctx context.Context, policyID bson.ObjectID, sort, researchType string, page, limit int, userID *bson.ObjectID) ([]*models.ResearchResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	filter := bson.M{"policyId": policyID}
	if researchType != "" {
		filter["type"] = researchType
	}

	total, err := r.research.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	sortDoc := r.getSortDoc(sort)
	findOpts := options.Find().
		SetSort(sortDoc).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, err := r.research.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var items []models.Research
	if err := cursor.All(ctx, &items); err != nil {
		return nil, 0, err
	}

	responses := make([]*models.ResearchResponse, len(items))
	for i := range items {
		resp, _ := r.toResponse(ctx, &items[i], userID)
		responses[i] = resp
	}

	return responses, total, nil
}

func (r *ResearchRepository) GetByID(ctx context.Context, policyID, researchID bson.ObjectID, userID *bson.ObjectID) (*models.ResearchResponse, error) {
	var research models.Research
	err := r.research.FindOne(ctx, bson.M{"_id": researchID, "policyId": policyID}).Decode(&research)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.toResponse(ctx, &research, userID)
}

func (r *ResearchRepository) Update(ctx context.Context, researchID, authorID bson.ObjectID, title, content string, sources []models.Source) (*models.ResearchResponse, error) {
	// Re-detect institutional for sources
	for i := range sources {
		sources[i].Institutional = r.IsInstitutional(sources[i].URL)
	}

	now := time.Now().UTC()
	result, err := r.research.UpdateOne(ctx,
		bson.M{"_id": researchID, "authorId": authorID},
		bson.M{"$set": bson.M{
			"title":     title,
			"content":   content,
			"sources":   sources,
			"updatedAt": now,
		}},
	)
	if err != nil {
		return nil, err
	}
	if result.MatchedCount == 0 {
		return nil, mongo.ErrNoDocuments
	}

	var research models.Research
	if err := r.research.FindOne(ctx, bson.M{"_id": researchID}).Decode(&research); err != nil {
		return nil, err
	}

	return r.toResponse(ctx, &research, nil)
}

func (r *ResearchRepository) React(ctx context.Context, userID, researchID bson.ObjectID, value int) error {
	// Get existing reaction
	var existing models.Reaction
	err := r.reactions.FindOne(ctx, bson.M{
		"userId":     userID,
		"entityId":   researchID,
		"entityType": "research",
	}).Decode(&existing)

	oldValue := 0
	if err == nil {
		oldValue = existing.Value
	}

	if value == 0 {
		// Remove reaction
		_, _ = r.reactions.DeleteOne(ctx, bson.M{
			"userId":     userID,
			"entityId":   researchID,
			"entityType": "research",
		})
	} else {
		// Upsert reaction
		_, err = r.reactions.UpdateOne(ctx,
			bson.M{"userId": userID, "entityId": researchID, "entityType": "research"},
			bson.M{"$set": bson.M{
				"userId":     userID,
				"entityId":   researchID,
				"entityType": "research",
				"value":      value,
				"createdAt":  time.Now().UTC(),
			}},
			options.UpdateOne().SetUpsert(true),
		)
		if err != nil {
			return err
		}
	}

	// Update research counters
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
		_, err = r.research.UpdateOne(ctx,
			bson.M{"_id": researchID},
			bson.M{"$inc": inc},
		)
		if err != nil {
			return err
		}
	}

	// Recalculate score
	var research models.Research
	if err := r.research.FindOne(ctx, bson.M{"_id": researchID}).Decode(&research); err == nil {
		score := research.Upvotes - research.Downvotes
		_, _ = r.research.UpdateOne(ctx,
			bson.M{"_id": researchID},
			bson.M{"$set": bson.M{"score": score}},
		)
	}

	return nil
}

func (r *ResearchRepository) GetUserVote(ctx context.Context, userID, entityID bson.ObjectID) int {
	var reaction models.Reaction
	err := r.reactions.FindOne(ctx, bson.M{
		"userId":     userID,
		"entityId":   entityID,
		"entityType": "research",
	}).Decode(&reaction)
	if err != nil {
		return 0
	}
	return reaction.Value
}

func (r *ResearchRepository) toResponse(ctx context.Context, research *models.Research, userID *bson.ObjectID) (*models.ResearchResponse, error) {
	resp := &models.ResearchResponse{
		ID:         research.ID,
		PolicyID:   research.PolicyID,
		AuthorID:   research.AuthorID,
		AuthorType: research.AuthorType,
		Title:      research.Title,
		Type:       research.Type,
		Content:    research.Content,
		Sources:    research.Sources,
		Upvotes:    research.Upvotes,
		Downvotes:  research.Downvotes,
		Score:      research.Score,
		CitedBy:    research.CitedBy,
		CreatedAt:  research.CreatedAt,
		UpdatedAt:  research.UpdatedAt,
	}

	r.enrichResponse(ctx, resp)

	if userID != nil {
		resp.UserVote = r.GetUserVote(ctx, *userID, research.ID)
	}

	return resp, nil
}

func (r *ResearchRepository) enrichResponse(ctx context.Context, resp *models.ResearchResponse) {
	var user models.User
	if err := r.users.FindOne(ctx, bson.M{"_id": resp.AuthorID}).Decode(&user); err == nil {
		resp.AuthorUsername = user.Username
		resp.AuthorRepTier = user.Reputation.Tier
	}
}

func (r *ResearchRepository) getSortDoc(sort string) bson.D {
	switch sort {
	case "newest":
		return bson.D{{Key: "createdAt", Value: -1}}
	case "top":
		return bson.D{{Key: "score", Value: -1}, {Key: "createdAt", Value: -1}}
	case "most_cited":
		return bson.D{{Key: "citedBy", Value: -1}, {Key: "createdAt", Value: -1}}
	default:
		return bson.D{{Key: "createdAt", Value: -1}}
	}
}
