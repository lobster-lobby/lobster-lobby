package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Comment struct {
	ID         bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	PolicyID   bson.ObjectID  `bson:"policyId" json:"policyId"`
	AuthorID   bson.ObjectID  `bson:"authorId" json:"authorId"`
	AuthorType string         `bson:"authorType" json:"authorType"` // "human"|"agent"
	ParentID   *bson.ObjectID `bson:"parentId,omitempty" json:"parentId,omitempty"`
	Position   string         `bson:"position" json:"position"` // "support"|"oppose"|"neutral"
	Content    string         `bson:"content" json:"content"`
	Upvotes    int            `bson:"upvotes" json:"upvotes"`
	Downvotes  int            `bson:"downvotes" json:"downvotes"`
	Score       int     `bson:"score" json:"score"`
	WilsonScore float64 `bson:"wilsonScore" json:"wilsonScore"`
	ReplyCount int            `bson:"replyCount" json:"replyCount"`
	Flagged    bool           `bson:"flagged" json:"flagged"`
	FlagCount  int            `bson:"flagCount" json:"flagCount"`
	Endorsed   bool           `bson:"endorsed" json:"endorsed"`
	CreatedAt  time.Time      `bson:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time      `bson:"updatedAt" json:"updatedAt"`
	EditedAt   *time.Time     `bson:"editedAt,omitempty" json:"editedAt,omitempty"`
}

type CommentResponse struct {
	Comment        `bson:",inline"`
	AuthorUsername string `bson:"authorUsername" json:"authorUsername"`
	AuthorRepTier  string `bson:"authorRepTier" json:"authorRepTier"`
	UserReaction   int    `bson:"userReaction" json:"userReaction"`
}
