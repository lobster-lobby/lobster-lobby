package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// CampaignActivityType represents the type of user activity on a campaign.
type CampaignActivityType string

const (
	CampaignActivityJoin    CampaignActivityType = "join"
	CampaignActivityShare   CampaignActivityType = "share"
	CampaignActivityComment CampaignActivityType = "comment"
	CampaignActivityUpload  CampaignActivityType = "upload"
)

// ValidActivityTypes is a set of valid campaign activity types.
var ValidActivityTypes = map[CampaignActivityType]bool{
	CampaignActivityJoin:    true,
	CampaignActivityShare:   true,
	CampaignActivityComment: true,
	CampaignActivityUpload:  true,
}

// CampaignActivity represents a user action on a campaign.
type CampaignActivity struct {
	ID          bson.ObjectID        `bson:"_id,omitempty" json:"id"`
	Type        CampaignActivityType `bson:"type" json:"type"`
	UserID      bson.ObjectID        `bson:"userId" json:"userId"`
	CampaignID  bson.ObjectID        `bson:"campaignId" json:"campaignId"`
	Description string               `bson:"description" json:"description"`
	CreatedAt   time.Time            `bson:"createdAt" json:"createdAt"`
}
