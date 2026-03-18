package models

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// CampaignComment represents a comment or reply on a campaign discussion.
type CampaignComment struct {
	ID         bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	CampaignID bson.ObjectID  `bson:"campaignId" json:"campaignId"`
	ParentID   *bson.ObjectID `bson:"parentId,omitempty" json:"parentId,omitempty"`
	AuthorID   bson.ObjectID  `bson:"authorId" json:"authorId"`
	AuthorName string         `bson:"authorName" json:"authorName"`
	Body       string         `bson:"body" json:"body"`
	Pinned     bool           `bson:"pinned" json:"pinned"`
	Votes      int            `bson:"votes" json:"votes"`
	CreatedAt  time.Time      `bson:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time      `bson:"updatedAt" json:"updatedAt"`
}

// Validate checks that the comment has valid field values.
func (c *CampaignComment) Validate() error {
	if len(c.Body) < 1 || len(c.Body) > 2000 {
		return errors.New("body must be between 1 and 2000 characters")
	}
	if c.CampaignID.IsZero() {
		return errors.New("campaignId is required")
	}
	if c.AuthorID.IsZero() {
		return errors.New("authorId is required")
	}
	if c.AuthorName == "" {
		return errors.New("authorName is required")
	}
	return nil
}

// CampaignCommentVote tracks a user's vote on a campaign comment.
type CampaignCommentVote struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	CommentID bson.ObjectID `bson:"commentId" json:"commentId"`
	UserID    bson.ObjectID `bson:"userId" json:"userId"`
	Value     int           `bson:"value" json:"value"` // 1 for upvote, -1 for downvote
	CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time     `bson:"updatedAt" json:"updatedAt"`
}
