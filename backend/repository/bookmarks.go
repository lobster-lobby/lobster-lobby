package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/lobster-lobby/lobster-lobby/models"
)

// ToggleBookmark adds a policy to the user's bookmarks if not present, or removes it if already bookmarked.
// Returns true if the policy is now bookmarked, false if it was removed.
func (r *UserRepository) ToggleBookmark(ctx context.Context, userID, policyID bson.ObjectID) (bool, error) {
	user, err := r.FindByID(ctx, userID)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, nil
	}

	for _, id := range user.Bookmarks {
		if id == policyID {
			// Remove bookmark
			_, err := r.coll.UpdateOne(ctx,
				bson.M{"_id": userID},
				bson.M{
					"$pull": bson.M{"bookmarks": policyID},
					"$set":  bson.M{"updatedAt": time.Now().UTC()},
				},
			)
			return false, err
		}
	}

	// Add bookmark
	_, err = r.coll.UpdateOne(ctx,
		bson.M{"_id": userID},
		bson.M{
			"$addToSet": bson.M{"bookmarks": policyID},
			"$set":      bson.M{"updatedAt": time.Now().UTC()},
		},
	)
	return true, err
}

// FindBookmarkedPolicies returns a paginated list of the user's bookmarked policies.
func (r *UserRepository) FindBookmarkedPolicies(ctx context.Context, userID bson.ObjectID, page, limit int) ([]models.Policy, int64, error) {
	user, err := r.FindByID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	if user == nil || len(user.Bookmarks) == 0 {
		return []models.Policy{}, 0, nil
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	total := int64(len(user.Bookmarks))

	// Reverse bookmarks so most recently added is first
	ids := make([]bson.ObjectID, len(user.Bookmarks))
	for i, id := range user.Bookmarks {
		ids[len(user.Bookmarks)-1-i] = id
	}

	// Apply pagination to the IDs
	start := (page - 1) * limit
	if int64(start) >= total {
		return []models.Policy{}, total, nil
	}
	end := start + limit
	if int64(end) > total {
		end = int(total)
	}
	pageIDs := ids[start:end]

	// Fetch policies by IDs, preserving order
	policyColl := r.coll.Database().Collection("policies")
	filter := bson.M{
		"_id":    bson.M{"$in": pageIDs},
		"status": bson.M{"$ne": models.PolicyStatusArchived},
	}
	cursor, err := policyColl.Find(ctx, filter, options.Find().SetLimit(int64(limit)))
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var fetched []models.Policy
	if err := cursor.All(ctx, &fetched); err != nil {
		return nil, 0, err
	}

	// Re-order to match pageIDs order
	policyMap := make(map[bson.ObjectID]models.Policy, len(fetched))
	for _, p := range fetched {
		policyMap[p.ID] = p
	}

	policies := make([]models.Policy, 0, len(pageIDs))
	for _, id := range pageIDs {
		if p, ok := policyMap[id]; ok {
			policies = append(policies, p)
		}
	}

	return policies, total, nil
}
