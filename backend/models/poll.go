package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Poll struct {
	ID          bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	PolicyID    bson.ObjectID   `bson:"policyId" json:"policyId"`
	AuthorID    bson.ObjectID   `bson:"authorId" json:"authorId"`
	AuthorName  string          `bson:"authorName" json:"authorName"`
	Question    string          `bson:"question" json:"question"`
	Options     []PollOption    `bson:"options" json:"options"`
	MultiSelect bool            `bson:"multiSelect" json:"multiSelect"`
	EndsAt      *time.Time      `bson:"endsAt,omitempty" json:"endsAt,omitempty"`
	Status      string          `bson:"status" json:"status"` // active, closed
	CreatedAt   time.Time       `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time       `bson:"updatedAt" json:"updatedAt"`
	TotalVotes  int             `bson:"totalVotes" json:"totalVotes"`
}

type PollOption struct {
	ID    bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Text  string        `bson:"text" json:"text"`
	Votes int           `bson:"votes" json:"votes"`
}

type PollVote struct {
	ID        bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	PollID    bson.ObjectID   `bson:"pollId" json:"pollId"`
	UserID    bson.ObjectID   `bson:"userId" json:"userId"`
	OptionIDs []bson.ObjectID `bson:"optionIds" json:"optionIds"`
	CreatedAt time.Time       `bson:"createdAt" json:"createdAt"`
}
