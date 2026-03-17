package models

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AssetType string

const (
	AssetTypeTextPost      AssetType = "text_post"
	AssetTypeEmailDraft    AssetType = "email_draft"
	AssetTypeSocialImage   AssetType = "social_image"
	AssetTypeInfographic   AssetType = "infographic"
	AssetTypeFlyer         AssetType = "flyer"
	AssetTypeLetter        AssetType = "letter"
	AssetTypeVideoScript   AssetType = "video_script"
	AssetTypeTalkingPoints AssetType = "talking_points"
)

var ValidAssetTypes = map[AssetType]bool{
	AssetTypeTextPost:      true,
	AssetTypeEmailDraft:    true,
	AssetTypeSocialImage:   true,
	AssetTypeInfographic:   true,
	AssetTypeFlyer:         true,
	AssetTypeLetter:        true,
	AssetTypeVideoScript:   true,
	AssetTypeTalkingPoints: true,
}

type CampaignAsset struct {
	ID                  bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	CampaignID          bson.ObjectID  `bson:"campaignId" json:"campaignId"`
	CreatedBy           bson.ObjectID  `bson:"createdBy" json:"createdBy"`
	CreatedByUsername   string         `bson:"createdByUsername" json:"createdByUsername"`
	Title               string         `bson:"title" json:"title"`
	Type                AssetType      `bson:"type" json:"type"`
	Content             string         `bson:"content,omitempty" json:"content,omitempty"`
	FileURL             string         `bson:"fileUrl,omitempty" json:"fileUrl,omitempty"`
	FileName            string         `bson:"fileName,omitempty" json:"fileName,omitempty"`
	FileSize            int64          `bson:"fileSize,omitempty" json:"fileSize,omitempty"`
	MimeType            string         `bson:"mimeType,omitempty" json:"mimeType,omitempty"`
	Description         string         `bson:"description" json:"description"`
	SubjectLine         string         `bson:"subjectLine,omitempty" json:"subjectLine,omitempty"`
	SuggestedRecipients string         `bson:"suggestedRecipients,omitempty" json:"suggestedRecipients,omitempty"`
	Upvotes             int            `bson:"upvotes" json:"upvotes"`
	Downvotes           int            `bson:"downvotes" json:"downvotes"`
	Score               int            `bson:"score" json:"score"`
	DownloadCount       int            `bson:"downloadCount" json:"downloadCount"`
	ShareCount          int            `bson:"shareCount" json:"shareCount"`
	SharesByPlatform    map[string]int `bson:"sharesByPlatform" json:"sharesByPlatform"`
	AIGenerated         bool           `bson:"aiGenerated" json:"aiGenerated"`
	CommentCount        int            `bson:"commentCount" json:"commentCount"`
	CreatedAt           time.Time      `bson:"createdAt" json:"createdAt"`
	UpdatedAt           time.Time      `bson:"updatedAt" json:"updatedAt"`
}

func (a *CampaignAsset) Validate() error {
	if len(a.Title) < 3 || len(a.Title) > 200 {
		return errors.New("title must be between 3 and 200 characters")
	}

	if !ValidAssetTypes[a.Type] {
		return errors.New("invalid asset type")
	}

	if len(a.Description) > 1000 {
		return errors.New("description must be at most 1000 characters")
	}

	// Text-based assets require content
	if a.IsTextBased() {
		if len(a.Content) < 10 {
			return errors.New("content must be at least 10 characters for text-based assets")
		}
		if len(a.Content) > 50000 {
			return errors.New("content must be at most 50000 characters")
		}
	}

	// Email drafts should have subject line
	if a.Type == AssetTypeEmailDraft && len(a.SubjectLine) == 0 {
		return errors.New("email drafts require a subject line")
	}

	return nil
}

func (a *CampaignAsset) IsTextBased() bool {
	switch a.Type {
	case AssetTypeTextPost, AssetTypeEmailDraft, AssetTypeLetter, AssetTypeVideoScript, AssetTypeTalkingPoints:
		return true
	default:
		return false
	}
}

func (a *CampaignAsset) IsFileBased() bool {
	return !a.IsTextBased()
}

// AssetVote tracks user votes on assets
type AssetVote struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	AssetID   bson.ObjectID `bson:"assetId" json:"assetId"`
	UserID    bson.ObjectID `bson:"userId" json:"userId"`
	Value     int           `bson:"value" json:"value"` // 1 for upvote, -1 for downvote
	CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time     `bson:"updatedAt" json:"updatedAt"`
}
