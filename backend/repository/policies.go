package repository

import (
	"context"
	"math"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/lobster-lobby/lobster-lobby/models"
)

type PolicyRepository struct {
	coll *mongo.Collection
}

func NewPolicyRepository(db *MongoDB) *PolicyRepository {
	return &PolicyRepository{coll: db.Database.Collection("policies")}
}

type PolicyListOpts struct {
	Page      int
	PerPage   int
	Sort      string
	Type      string
	Level     string
	State     string
	Tags      []string
	CreatedBy string
}

func (r *PolicyRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "slug", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "hotScore", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "createdAt", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "type", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "level", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "state", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "tags", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "createdBy", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
	})
	return err
}

func (r *PolicyRepository) Create(ctx context.Context, policy *models.Policy) error {
	now := time.Now().UTC()
	policy.CreatedAt = now
	policy.UpdatedAt = now
	if policy.ID.IsZero() {
		policy.ID = bson.NewObjectID()
	}
	if policy.Tags == nil {
		policy.Tags = []string{}
	}
	if policy.LinkedPolicies == nil {
		policy.LinkedPolicies = []bson.ObjectID{}
	}

	policy.HotScore = r.calculateHotScore(policy)

	_, err := r.coll.InsertOne(ctx, policy)
	return err
}

func (r *PolicyRepository) FindByID(ctx context.Context, id bson.ObjectID) (*models.Policy, error) {
	var p models.Policy
	err := r.coll.FindOne(ctx, bson.M{"_id": id, "status": bson.M{"$ne": models.PolicyStatusArchived}}).Decode(&p)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &p, err
}

func (r *PolicyRepository) FindBySlug(ctx context.Context, slug string) (*models.Policy, error) {
	var p models.Policy
	err := r.coll.FindOne(ctx, bson.M{"slug": slug, "status": bson.M{"$ne": models.PolicyStatusArchived}}).Decode(&p)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &p, err
}

func (r *PolicyRepository) Update(ctx context.Context, id bson.ObjectID, updates bson.M) error {
	updates["updatedAt"] = time.Now().UTC()
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": updates})
	return err
}

func (r *PolicyRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	return r.Update(ctx, id, bson.M{"status": models.PolicyStatusArchived})
}

func (r *PolicyRepository) List(ctx context.Context, opts PolicyListOpts) ([]models.Policy, int64, error) {
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PerPage < 1 || opts.PerPage > 100 {
		opts.PerPage = 20
	}

	filter := bson.M{"status": bson.M{"$ne": models.PolicyStatusArchived}}

	if opts.Type != "" {
		filter["type"] = opts.Type
	}
	if opts.Level != "" {
		filter["level"] = opts.Level
	}
	if opts.State != "" {
		filter["state"] = opts.State
	}
	if len(opts.Tags) > 0 {
		filter["tags"] = bson.M{"$in": opts.Tags}
	}
	if opts.CreatedBy != "" {
		if oid, err := bson.ObjectIDFromHex(opts.CreatedBy); err == nil {
			filter["createdBy"] = oid
		}
	}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	sortField := r.getSortField(opts.Sort)
	findOpts := options.Find().
		SetSort(sortField).
		SetSkip(int64((opts.Page - 1) * opts.PerPage)).
		SetLimit(int64(opts.PerPage))

	cursor, err := r.coll.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var policies []models.Policy
	if err := cursor.All(ctx, &policies); err != nil {
		return nil, 0, err
	}

	if policies == nil {
		policies = []models.Policy{}
	}

	return policies, total, nil
}

func (r *PolicyRepository) IncrementEngagement(ctx context.Context, id bson.ObjectID, field string) error {
	validFields := map[string]bool{
		"engagement.debateCount":   true,
		"engagement.researchCount": true,
		"engagement.pollCount":     true,
		"engagement.bookmarkCount": true,
		"engagement.viewCount":     true,
	}
	if !validFields[field] {
		return nil
	}

	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$inc": bson.M{field: 1},
			"$set": bson.M{"updatedAt": time.Now().UTC()},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var p models.Policy
	if err := result.Decode(&p); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		return err
	}

	newScore := r.calculateHotScore(&p)
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"hotScore": newScore}})
	return err
}

func (r *PolicyRepository) DecrementEngagement(ctx context.Context, id bson.ObjectID, field string) error {
	validFields := map[string]bool{
		"engagement.debateCount":   true,
		"engagement.researchCount": true,
		"engagement.pollCount":     true,
		"engagement.bookmarkCount": true,
		"engagement.viewCount":     true,
	}
	if !validFields[field] {
		return nil
	}

	result := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$inc": bson.M{field: -1},
			"$set": bson.M{"updatedAt": time.Now().UTC()},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var p models.Policy
	if err := result.Decode(&p); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		return err
	}

	newScore := r.calculateHotScore(&p)
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"hotScore": newScore}})
	return err
}

// AddLinkedPolicy adds a linked policy ID to a policy's linkedPolicies array (idempotent).
func (r *PolicyRepository) AddLinkedPolicy(ctx context.Context, policyID, linkedID bson.ObjectID) error {
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": policyID}, bson.M{
		"$addToSet": bson.M{"linkedPolicies": linkedID},
		"$set":      bson.M{"updatedAt": time.Now().UTC()},
	})
	return err
}

func (r *PolicyRepository) GenerateSlug(ctx context.Context, title string) (string, error) {
	baseSlug := toKebabCase(title)
	if len(baseSlug) > 100 {
		baseSlug = baseSlug[:100]
	}

	slug := baseSlug
	suffix := 2

	for {
		exists, err := r.slugExists(ctx, slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = baseSlug + "-" + itoa(suffix)
		suffix++
	}
}

func (r *PolicyRepository) slugExists(ctx context.Context, slug string) (bool, error) {
	count, err := r.coll.CountDocuments(ctx, bson.M{"slug": slug})
	return count > 0, err
}

func (r *PolicyRepository) calculateHotScore(p *models.Policy) float64 {
	engagementTotal := float64(p.Engagement.DebateCount + p.Engagement.ResearchCount*2 + p.Engagement.BookmarkCount)
	logScore := math.Log10(math.Max(engagementTotal, 1))
	timeScore := float64(p.CreatedAt.Unix()) / 45000.0
	return logScore + timeScore
}

func (r *PolicyRepository) getSortField(sort string) bson.D {
	now := time.Now().UTC()

	switch sort {
	case "new":
		return bson.D{{Key: "createdAt", Value: -1}}
	case "top_week":
		return bson.D{{Key: "createdAt", Value: -1}}
	case "top_month":
		return bson.D{{Key: "createdAt", Value: -1}}
	case "top_all":
		return bson.D{{Key: "engagement.viewCount", Value: -1}, {Key: "createdAt", Value: -1}}
	case "most_debated":
		return bson.D{{Key: "engagement.debateCount", Value: -1}, {Key: "createdAt", Value: -1}}
	case "hot":
		fallthrough
	default:
		_ = now
		return bson.D{{Key: "hotScore", Value: -1}}
	}
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func toKebabCase(s string) string {
	s = strings.ToLower(s)
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
