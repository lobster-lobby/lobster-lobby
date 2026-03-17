package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/lobster-lobby/lobster-lobby/models"
)

type AssetRepository struct {
	assets *mongo.Collection
	votes  *mongo.Collection
}

func NewAssetRepository(db *MongoDB) *AssetRepository {
	return &AssetRepository{
		assets: db.Database.Collection("campaign_assets"),
		votes:  db.Database.Collection("asset_votes"),
	}
}

func (r *AssetRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.assets.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "campaignId", Value: 1}}},
		{Keys: bson.D{{Key: "campaignId", Value: 1}, {Key: "type", Value: 1}}},
		{Keys: bson.D{{Key: "createdBy", Value: 1}}},
		{Keys: bson.D{{Key: "score", Value: -1}}},
		{Keys: bson.D{{Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "downloadCount", Value: -1}}},
		{Keys: bson.D{{Key: "shareCount", Value: -1}}},
	})
	if err != nil {
		return err
	}

	_, err = r.votes.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "assetId", Value: 1}, {Key: "userId", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	return err
}

type AssetListOpts struct {
	CampaignID string
	Page       int
	PerPage    int
	Sort       string // "top", "newest", "most_downloaded", "most_shared"
	Type       string // filter by asset type
}

func (r *AssetRepository) Create(ctx context.Context, asset *models.CampaignAsset) error {
	now := time.Now().UTC()
	asset.CreatedAt = now
	asset.UpdatedAt = now
	if asset.ID.IsZero() {
		asset.ID = bson.NewObjectID()
	}
	if asset.SharesByPlatform == nil {
		asset.SharesByPlatform = make(map[string]int)
	}

	_, err := r.assets.InsertOne(ctx, asset)
	return err
}

func (r *AssetRepository) GetByID(ctx context.Context, id string) (*models.CampaignAsset, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var asset models.CampaignAsset
	err = r.assets.FindOne(ctx, bson.M{"_id": oid}).Decode(&asset)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &asset, err
}

func (r *AssetRepository) List(ctx context.Context, opts AssetListOpts) ([]models.CampaignAsset, int64, error) {
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PerPage < 1 || opts.PerPage > 100 {
		opts.PerPage = 20
	}

	filter := bson.M{}

	if opts.CampaignID != "" {
		oid, err := bson.ObjectIDFromHex(opts.CampaignID)
		if err != nil {
			return nil, 0, err
		}
		filter["campaignId"] = oid
	}

	if opts.Type != "" {
		filter["type"] = opts.Type
	}

	total, err := r.assets.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	sortField := r.getSortField(opts.Sort)
	findOpts := options.Find().
		SetSort(sortField).
		SetSkip(int64((opts.Page - 1) * opts.PerPage)).
		SetLimit(int64(opts.PerPage))

	cursor, err := r.assets.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var assets []models.CampaignAsset
	if err := cursor.All(ctx, &assets); err != nil {
		return nil, 0, err
	}

	if assets == nil {
		assets = []models.CampaignAsset{}
	}

	return assets, total, nil
}

func (r *AssetRepository) Update(ctx context.Context, id string, updates bson.M) (*models.CampaignAsset, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	updates["updatedAt"] = time.Now().UTC()
	result := r.assets.FindOneAndUpdate(
		ctx,
		bson.M{"_id": oid},
		bson.M{"$set": updates},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var asset models.CampaignAsset
	if err := result.Decode(&asset); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &asset, nil
}

func (r *AssetRepository) Delete(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.assets.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

func (r *AssetRepository) IncrementDownload(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.assets.UpdateOne(
		ctx,
		bson.M{"_id": oid},
		bson.M{
			"$inc": bson.M{"downloadCount": 1},
			"$set": bson.M{"updatedAt": time.Now().UTC()},
		},
	)
	return err
}

func (r *AssetRepository) IncrementShare(ctx context.Context, id string, platform string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.assets.UpdateOne(
		ctx,
		bson.M{"_id": oid},
		bson.M{
			"$inc": bson.M{
				"shareCount":                       1,
				"sharesByPlatform." + platform:     1,
			},
			"$set": bson.M{"updatedAt": time.Now().UTC()},
		},
	)
	return err
}

// Vote handling
func (r *AssetRepository) GetVote(ctx context.Context, assetID, userID string) (*models.AssetVote, error) {
	assetOID, err := bson.ObjectIDFromHex(assetID)
	if err != nil {
		return nil, err
	}
	userOID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var vote models.AssetVote
	err = r.votes.FindOne(ctx, bson.M{
		"assetId": assetOID,
		"userId":  userOID,
	}).Decode(&vote)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &vote, err
}

func (r *AssetRepository) SetVote(ctx context.Context, assetID, userID string, value int) error {
	assetOID, err := bson.ObjectIDFromHex(assetID)
	if err != nil {
		return err
	}
	userOID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	// Get existing vote to calculate delta
	existingVote, _ := r.GetVote(ctx, assetID, userID)
	oldValue := 0
	if existingVote != nil {
		oldValue = existingVote.Value
	}

	if value == 0 {
		// Remove vote
		_, err = r.votes.DeleteOne(ctx, bson.M{
			"assetId": assetOID,
			"userId":  userOID,
		})
	} else {
		// Upsert vote
		_, err = r.votes.UpdateOne(
			ctx,
			bson.M{"assetId": assetOID, "userId": userOID},
			bson.M{
				"$set": bson.M{
					"value":     value,
					"updatedAt": now,
				},
				"$setOnInsert": bson.M{
					"_id":       bson.NewObjectID(),
					"assetId":   assetOID,
					"userId":    userOID,
					"createdAt": now,
				},
			},
			options.UpdateOne().SetUpsert(true),
		)
	}
	if err != nil {
		return err
	}

	// Calculate deltas
	upvoteDelta := 0
	downvoteDelta := 0

	// Remove old vote effect
	if oldValue == 1 {
		upvoteDelta--
	} else if oldValue == -1 {
		downvoteDelta--
	}

	// Add new vote effect
	if value == 1 {
		upvoteDelta++
	} else if value == -1 {
		downvoteDelta++
	}

	// Update asset counts
	if upvoteDelta != 0 || downvoteDelta != 0 {
		scoreDelta := upvoteDelta - downvoteDelta
		_, err = r.assets.UpdateOne(
			ctx,
			bson.M{"_id": assetOID},
			bson.M{
				"$inc": bson.M{
					"upvotes":   upvoteDelta,
					"downvotes": downvoteDelta,
					"score":     scoreDelta,
				},
				"$set": bson.M{"updatedAt": now},
			},
		)
	}

	return err
}

func (r *AssetRepository) CountByCampaign(ctx context.Context, campaignID string) (int64, error) {
	oid, err := bson.ObjectIDFromHex(campaignID)
	if err != nil {
		return 0, err
	}

	return r.assets.CountDocuments(ctx, bson.M{"campaignId": oid})
}

// GetBatchVotes returns a map of assetID → vote value for the given user and asset IDs.
func (r *AssetRepository) GetBatchVotes(ctx context.Context, assetIDs []string, userID string) (map[string]int, error) {
	userOID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	oids := make([]bson.ObjectID, 0, len(assetIDs))
	for _, id := range assetIDs {
		oid, err := bson.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		oids = append(oids, oid)
	}

	cursor, err := r.votes.Find(ctx, bson.M{
		"assetId": bson.M{"$in": oids},
		"userId":  userOID,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]int)
	for cursor.Next(ctx) {
		var vote models.AssetVote
		if err := cursor.Decode(&vote); err != nil {
			continue
		}
		result[vote.AssetID.Hex()] = vote.Value
	}
	return result, cursor.Err()
}

func (r *AssetRepository) getSortField(sort string) bson.D {
	switch sort {
	case "newest":
		return bson.D{{Key: "createdAt", Value: -1}}
	case "most_downloaded":
		return bson.D{{Key: "downloadCount", Value: -1}, {Key: "createdAt", Value: -1}}
	case "most_shared":
		return bson.D{{Key: "shareCount", Value: -1}, {Key: "createdAt", Value: -1}}
	case "top":
		fallthrough
	default:
		return bson.D{{Key: "score", Value: -1}, {Key: "createdAt", Value: -1}}
	}
}
