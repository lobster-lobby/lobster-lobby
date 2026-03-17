package models

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CampaignStatus string

const (
	CampaignStatusActive    CampaignStatus = "active"
	CampaignStatusPaused    CampaignStatus = "paused"
	CampaignStatusCompleted CampaignStatus = "completed"
	CampaignStatusArchived  CampaignStatus = "archived"
)

type Milestone struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string        `bson:"title" json:"title"`
	Description string        `bson:"description" json:"description"`
	Target      int           `bson:"target" json:"target"`
	Current     int           `bson:"current" json:"current"`
	Completed   bool          `bson:"completed" json:"completed"`
	CompletedAt *time.Time    `bson:"completedAt,omitempty" json:"completedAt,omitempty"`
	CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time     `bson:"updatedAt" json:"updatedAt"`
}

type CampaignMetrics struct {
	TotalDownloads     int            `bson:"totalDownloads" json:"totalDownloads"`
	TotalShares        int            `bson:"totalShares" json:"totalShares"`
	SharesByPlatform   map[string]int `bson:"sharesByPlatform" json:"sharesByPlatform"`
	UniqueParticipants int            `bson:"uniqueParticipants" json:"uniqueParticipants"`
	AssetCount         int            `bson:"assetCount" json:"assetCount"`
	CommentCount       int            `bson:"commentCount" json:"commentCount"`
}

type Campaign struct {
	ID                bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	Title             string          `bson:"title" json:"title"`
	Slug              string          `bson:"slug" json:"slug"`
	PolicyID          bson.ObjectID   `bson:"policyId" json:"policyId"`
	CreatedBy         bson.ObjectID   `bson:"createdBy" json:"createdBy"`
	Objective         string          `bson:"objective" json:"objective"`
	Target            string          `bson:"target" json:"target"`
	Description       string          `bson:"description" json:"description"`
	Status            CampaignStatus  `bson:"status" json:"status"`
	Milestones        []Milestone     `bson:"milestones" json:"milestones"`
	CompletionSummary string          `bson:"completionSummary" json:"completionSummary"`
	Metrics           CampaignMetrics `bson:"metrics" json:"metrics"`
	TrendingScore     float64         `bson:"trendingScore" json:"trendingScore"`
	CreatedAt         time.Time       `bson:"createdAt" json:"createdAt"`
	UpdatedAt         time.Time       `bson:"updatedAt" json:"updatedAt"`
}

func (c *Campaign) Validate() error {
	if len(c.Title) < 5 || len(c.Title) > 200 {
		return errors.New("title must be between 5 and 200 characters")
	}

	if len(c.Objective) < 10 || len(c.Objective) > 500 {
		return errors.New("objective must be between 10 and 500 characters")
	}

	if len(c.Target) < 5 || len(c.Target) > 500 {
		return errors.New("target must be between 5 and 500 characters")
	}

	if len(c.Description) < 20 || len(c.Description) > 5000 {
		return errors.New("description must be between 20 and 5000 characters")
	}

	if c.Status == "" {
		c.Status = CampaignStatusActive
	} else if c.Status != CampaignStatusActive && c.Status != CampaignStatusPaused &&
		c.Status != CampaignStatusCompleted && c.Status != CampaignStatusArchived {
		return errors.New("status must be one of: active, paused, completed, archived")
	}

	return nil
}

// CalculateTrendingScore computes a trending score based on recent activity metrics.
// Higher share/participant counts and recency increase the score.
func (c *Campaign) CalculateTrendingScore() float64 {
	ageHours := time.Since(c.CreatedAt).Hours()
	if ageHours < 1 {
		ageHours = 1
	}
	activity := float64(c.Metrics.TotalShares*3+c.Metrics.UniqueParticipants*2+c.Metrics.CommentCount)
	return activity / ageHours
}

var slugNonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func GenerateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = slugNonAlphanumeric.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > 100 {
		slug = strings.TrimRight(slug[:100], "-")
	}
	return slug
}
