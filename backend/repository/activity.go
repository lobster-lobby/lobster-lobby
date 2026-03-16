package repository

import (
	"context"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ActivityItem struct {
	Type        string    `json:"type"`
	PolicyID    string    `json:"policyId"`
	PolicyTitle string    `json:"policyTitle"`
	CreatedAt   time.Time `json:"createdAt"`
}

type UserStats struct {
	PoliciesCreated   int64 `json:"policiesCreated"`
	DebateComments    int64 `json:"debateComments"`
	ResearchSubmitted int64 `json:"researchSubmitted"`
	Bookmarks         int   `json:"bookmarks"`
}

type ActivityRepository struct {
	db *mongo.Database
}

func NewActivityRepository(mongoDB *MongoDB) *ActivityRepository {
	return &ActivityRepository{db: mongoDB.Database}
}

func (r *ActivityRepository) GetUserStats(ctx context.Context, userID bson.ObjectID, bookmarkCount int) (*UserStats, error) {
	stats := &UserStats{Bookmarks: bookmarkCount}

	policyColl := r.db.Collection("policies")
	count, err := policyColl.CountDocuments(ctx, bson.M{
		"createdBy": userID,
		"status":    bson.M{"$ne": "archived"},
	})
	if err != nil {
		return nil, err
	}
	stats.PoliciesCreated = count

	commentColl := r.db.Collection("comments")
	count, err = commentColl.CountDocuments(ctx, bson.M{"authorId": userID})
	if err != nil {
		// Collection may not exist; treat as zero
		stats.DebateComments = 0
	} else {
		stats.DebateComments = count
	}

	researchColl := r.db.Collection("research")
	count, err = researchColl.CountDocuments(ctx, bson.M{"submittedBy": userID})
	if err != nil {
		stats.ResearchSubmitted = 0
	} else {
		stats.ResearchSubmitted = count
	}

	return stats, nil
}

func (r *ActivityRepository) GetRecentActivity(ctx context.Context, userID bson.ObjectID, limit int) ([]ActivityItem, error) {
	var items []ActivityItem

	// Policies created by user
	policyColl := r.db.Collection("policies")
	findOpts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetLimit(int64(limit)).
		SetProjection(bson.M{"_id": 1, "title": 1, "createdAt": 1})

	cursor, err := policyColl.Find(ctx, bson.M{
		"createdBy": userID,
		"status":    bson.M{"$ne": "archived"},
	}, findOpts)
	if err == nil {
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var doc struct {
				ID        bson.ObjectID `bson:"_id"`
				Title     string        `bson:"title"`
				CreatedAt time.Time     `bson:"createdAt"`
			}
			if err := cursor.Decode(&doc); err == nil {
				items = append(items, ActivityItem{
					Type:        "policy_created",
					PolicyID:    doc.ID.Hex(),
					PolicyTitle: doc.Title,
					CreatedAt:   doc.CreatedAt,
				})
			}
		}
	}

	// Comments posted by user
	commentColl := r.db.Collection("comments")
	commentOpts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetLimit(int64(limit)).
		SetProjection(bson.M{"_id": 1, "policyId": 1, "createdAt": 1})

	cursor, err = commentColl.Find(ctx, bson.M{"authorId": userID}, commentOpts)
	if err == nil {
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var doc struct {
				PolicyID  bson.ObjectID `bson:"policyId"`
				CreatedAt time.Time     `bson:"createdAt"`
			}
			if err := cursor.Decode(&doc); err == nil {
				items = append(items, ActivityItem{
					Type:      "comment_posted",
					PolicyID:  doc.PolicyID.Hex(),
					CreatedAt: doc.CreatedAt,
				})
			}
		}
	}

	// Research submitted by user
	researchColl := r.db.Collection("research")
	researchOpts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetLimit(int64(limit)).
		SetProjection(bson.M{"_id": 1, "policyId": 1, "title": 1, "createdAt": 1})

	cursor, err = researchColl.Find(ctx, bson.M{"submittedBy": userID}, researchOpts)
	if err == nil {
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var doc struct {
				PolicyID  bson.ObjectID `bson:"policyId"`
				Title     string        `bson:"title"`
				CreatedAt time.Time     `bson:"createdAt"`
			}
			if err := cursor.Decode(&doc); err == nil {
				items = append(items, ActivityItem{
					Type:        "research_submitted",
					PolicyID:    doc.PolicyID.Hex(),
					PolicyTitle: doc.Title,
					CreatedAt:   doc.CreatedAt,
				})
			}
		}
	}

	// Sort all items by createdAt desc
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	if len(items) > limit {
		items = items[:limit]
	}
	if items == nil {
		items = []ActivityItem{}
	}

	return items, nil
}
