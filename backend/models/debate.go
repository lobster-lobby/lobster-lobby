package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Debate struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Slug        string        `bson:"slug" json:"slug"`
	Title       string        `bson:"title" json:"title"`
	Description string        `bson:"description" json:"description"`
	CreatorID   bson.ObjectID `bson:"creatorId" json:"creatorId"`
	Status      string        `bson:"status" json:"status"` // "open"|"closed"
	CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time     `bson:"updatedAt" json:"updatedAt"`
}

type DebateResponse struct {
	Debate          `bson:",inline"`
	CreatorUsername string `bson:"creatorUsername" json:"creatorUsername"`
	ArgumentCount   int    `bson:"argumentCount" json:"argumentCount"`
}

type Argument struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	DebateID   bson.ObjectID `bson:"debateId" json:"debateId"`
	AuthorID   bson.ObjectID `bson:"authorId" json:"authorId"`
	AuthorType string        `bson:"authorType" json:"authorType"` // "human"|"agent"
	Side       string        `bson:"side" json:"side"`             // "pro"|"con"
	Content    string        `bson:"content" json:"content"`
	Upvotes    int           `bson:"upvotes" json:"upvotes"`
	Downvotes  int           `bson:"downvotes" json:"downvotes"`
	Score      int           `bson:"score" json:"score"` // upvotes - downvotes
	CreatedAt  time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time     `bson:"updatedAt" json:"updatedAt"`
}

type ArgumentResponse struct {
	Argument       `bson:",inline"`
	AuthorUsername string `bson:"authorUsername" json:"authorUsername"`
	AuthorRepTier  string `bson:"authorRepTier" json:"authorRepTier"`
	UserVote       int    `bson:"userVote" json:"userVote"` // 0, 1, or -1
}

type Vote struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	DebateID   bson.ObjectID `bson:"debateId" json:"debateId"`
	ArgumentID bson.ObjectID `bson:"argumentId" json:"argumentId"`
	UserID     bson.ObjectID `bson:"userId" json:"userId"`
	Value      int           `bson:"value" json:"value"` // +1 or -1
	CreatedAt  time.Time     `bson:"createdAt" json:"createdAt"`
}
