package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// CampaignEventType represents the type of campaign event.
type CampaignEventType string

const (
	CampaignEventCreated          CampaignEventType = "created"
	CampaignEventAssetAdded       CampaignEventType = "asset_added"
	CampaignEventMilestone        CampaignEventType = "milestone"
	CampaignEventStatusChange     CampaignEventType = "status_change"
	CampaignEventCommentMilestone CampaignEventType = "comment_milestone"
)

// CampaignEvent represents a timeline event for a campaign.
type CampaignEvent struct {
	ID          bson.ObjectID     `bson:"_id,omitempty" json:"id"`
	CampaignID  bson.ObjectID     `bson:"campaignId" json:"campaignId"`
	Type        CampaignEventType `bson:"type" json:"type"`
	Title       string            `bson:"title" json:"title"`
	Description string            `bson:"description" json:"description"`
	Metadata    map[string]any    `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt   time.Time         `bson:"createdAt" json:"createdAt"`
}

// ValidEventTypes is a set of valid campaign event types.
var ValidEventTypes = map[CampaignEventType]bool{
	CampaignEventCreated:          true,
	CampaignEventAssetAdded:       true,
	CampaignEventMilestone:        true,
	CampaignEventStatusChange:     true,
	CampaignEventCommentMilestone: true,
}
