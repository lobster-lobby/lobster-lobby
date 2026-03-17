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

type CampaignRepository struct {
	coll *mongo.Collection
}

func NewCampaignRepository(db *MongoDB) *CampaignRepository {
	return &CampaignRepository{coll: db.Database.Collection("campaigns")}
}

type CampaignListOpts struct {
	Page     int
	PerPage  int
	Sort     string // "trending", "newest", "participants", "shares"
	Status   string
	PolicyID string
}

func (r *CampaignRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "slug", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "policyId", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "trendingScore", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "createdAt", Value: -1}},
		},
	})
	return err
}

func (r *CampaignRepository) Create(ctx context.Context, campaign *models.Campaign) error {
	now := time.Now().UTC()
	campaign.CreatedAt = now
	campaign.UpdatedAt = now
	if campaign.ID.IsZero() {
		campaign.ID = bson.NewObjectID()
	}
	if campaign.Milestones == nil {
		campaign.Milestones = []models.Milestone{}
	}
	if campaign.Metrics.SharesByPlatform == nil {
		campaign.Metrics.SharesByPlatform = make(map[string]int)
	}

	// Generate unique slug
	baseSlug := models.GenerateSlug(campaign.Title)
	slug := baseSlug
	suffix := 2

	for {
		exists, err := r.slugExists(ctx, slug)
		if err != nil {
			return err
		}
		if !exists {
			campaign.Slug = slug
			break
		}
		slug = baseSlug + "-" + itoa(suffix)
		suffix++
	}

	campaign.TrendingScore = r.calculateTrendingScore(campaign)

	_, err := r.coll.InsertOne(ctx, campaign)
	return err
}

func (r *CampaignRepository) GetByID(ctx context.Context, id string) (*models.Campaign, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var c models.Campaign
	err = r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&c)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &c, err
}

func (r *CampaignRepository) FindBySlug(ctx context.Context, slug string) (*models.Campaign, error) {
	var c models.Campaign
	err := r.coll.FindOne(ctx, bson.M{"slug": slug}).Decode(&c)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &c, err
}

func (r *CampaignRepository) FindByPolicy(ctx context.Context, policyID string) ([]models.Campaign, error) {
	oid, err := bson.ObjectIDFromHex(policyID)
	if err != nil {
		return []models.Campaign{}, nil
	}

	filter := bson.M{"policyId": oid}
	cursor, err := r.coll.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var campaigns []models.Campaign
	if err := cursor.All(ctx, &campaigns); err != nil {
		return nil, err
	}

	if campaigns == nil {
		campaigns = []models.Campaign{}
	}

	return campaigns, nil
}

func (r *CampaignRepository) List(ctx context.Context, opts CampaignListOpts) ([]models.Campaign, int64, error) {
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PerPage < 1 || opts.PerPage > 100 {
		opts.PerPage = 20
	}

	filter := bson.M{}

	if opts.Status != "" {
		filter["status"] = opts.Status
	}
	if opts.PolicyID != "" {
		if oid, err := bson.ObjectIDFromHex(opts.PolicyID); err == nil {
			filter["policyId"] = oid
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

	var campaigns []models.Campaign
	if err := cursor.All(ctx, &campaigns); err != nil {
		return nil, 0, err
	}

	if campaigns == nil {
		campaigns = []models.Campaign{}
	}

	return campaigns, total, nil
}

func (r *CampaignRepository) Update(ctx context.Context, id string, updates bson.M) (*models.Campaign, error) {
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

	var c models.Campaign
	if err := result.Decode(&c); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &c, nil
}

func (r *CampaignRepository) UpdateStatus(ctx context.Context, id string, status models.CampaignStatus) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	updates := bson.M{
		"status":    status,
		"updatedAt": time.Now().UTC(),
	}

	_, err = r.coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": updates})
	return err
}

func (r *CampaignRepository) UpdateMetrics(ctx context.Context, id string, metrics models.CampaignMetrics) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	updates := bson.M{
		"metrics":   metrics,
		"updatedAt": time.Now().UTC(),
	}

	_, err = r.coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": updates})
	return err
}

// IncrMetric atomically increments a campaign metric field by delta using $inc.
func (r *CampaignRepository) IncrMetric(ctx context.Context, campaignID, field string, delta int) error {
	oid, err := bson.ObjectIDFromHex(campaignID)
	if err != nil {
		return err
	}
	_, err = r.coll.UpdateOne(
		ctx,
		bson.M{"_id": oid},
		bson.M{
			"$inc": bson.M{"metrics." + field: delta},
			"$set": bson.M{"updatedAt": time.Now().UTC()},
		},
	)
	return err
}

func (r *CampaignRepository) RecalcTrending(ctx context.Context, id string) error {
	campaign, err := r.GetByID(ctx, id)
	if err != nil || campaign == nil {
		return err
	}

	newScore := r.calculateTrendingScore(campaign)

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"trendingScore": newScore}})
	return err
}

func (r *CampaignRepository) slugExists(ctx context.Context, slug string) (bool, error) {
	count, err := r.coll.CountDocuments(ctx, bson.M{"slug": slug})
	return count > 0, err
}

func (r *CampaignRepository) calculateTrendingScore(c *models.Campaign) float64 {
	// Formula: (downloads_7d * 1 + shares_7d * 3 + new_assets_7d * 5 + comments_7d * 0.5) * recency_multiplier
	// Since we don't have 7d breakdown, use totals as approximation
	downloads := float64(c.Metrics.TotalDownloads)
	shares := float64(c.Metrics.TotalShares)
	assets := float64(c.Metrics.AssetCount)
	comments := float64(c.Metrics.CommentCount)

	activityScore := downloads*1 + shares*3 + assets*5 + comments*0.5

	// Recency multiplier: max(0, 1 - days_since_last_activity / 14)
	daysSinceUpdate := time.Since(c.UpdatedAt).Hours() / 24.0
	recencyMultiplier := math.Max(0, 1-daysSinceUpdate/14.0)

	return activityScore * recencyMultiplier
}

func (r *CampaignRepository) getSortField(sort string) bson.D {
	switch sort {
	case "newest":
		return bson.D{{Key: "createdAt", Value: -1}}
	case "participants":
		return bson.D{{Key: "metrics.uniqueParticipants", Value: -1}, {Key: "createdAt", Value: -1}}
	case "shares":
		return bson.D{{Key: "metrics.totalShares", Value: -1}, {Key: "createdAt", Value: -1}}
	case "trending":
		fallthrough
	default:
		return bson.D{{Key: "trendingScore", Value: -1}}
	}
}
