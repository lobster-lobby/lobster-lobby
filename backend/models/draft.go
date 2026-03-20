package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Draft struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	PolicyID     bson.ObjectID `bson:"policyId" json:"policyId"`
	AuthorID     bson.ObjectID `bson:"authorId" json:"authorId"`
	AuthorName   string        `bson:"authorName" json:"authorName"`
	Title        string        `bson:"title" json:"title"`
	Content      string        `bson:"content" json:"content"` // Markdown
	Category     string        `bson:"category" json:"category"` // amendment, talking-point, position-statement, full-text
	Status       string        `bson:"status" json:"status"` // draft, published, archived
	Endorsements int           `bson:"endorsements" json:"endorsements"`
	Version      int           `bson:"version" json:"version"`
	CreatedAt    time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time     `bson:"updatedAt" json:"updatedAt"`
}

type DraftEndorsement struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	DraftID   bson.ObjectID `bson:"draftId" json:"draftId"`
	UserID    bson.ObjectID `bson:"userId" json:"userId"`
	CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
}

type DraftComment struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	DraftID    bson.ObjectID `bson:"draftId" json:"draftId"`
	AuthorID   bson.ObjectID `bson:"authorId" json:"authorId"`
	AuthorName string        `bson:"authorName" json:"authorName"`
	Content    string        `bson:"content" json:"content"`
	CreatedAt  time.Time     `bson:"createdAt" json:"createdAt"`
}
