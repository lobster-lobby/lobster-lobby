package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Research struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	PolicyID   bson.ObjectID `bson:"policyId" json:"policyId"`
	AuthorID   bson.ObjectID `bson:"authorId" json:"authorId"`
	AuthorType string        `bson:"authorType" json:"authorType"`
	Title      string        `bson:"title" json:"title"`
	Type       string        `bson:"type" json:"type"` // "analysis"|"news"|"data"|"academic"|"government"
	Content    string        `bson:"content" json:"content"`
	Sources    []Source      `bson:"sources" json:"sources"`
	Upvotes    int           `bson:"upvotes" json:"upvotes"`
	Downvotes  int           `bson:"downvotes" json:"downvotes"`
	Score      int           `bson:"score" json:"score"`
	CitedBy    int           `bson:"citedBy" json:"citedBy"`
	CreatedAt  time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time     `bson:"updatedAt" json:"updatedAt"`
}

type Source struct {
	URL           string     `bson:"url" json:"url"`
	Title         string     `bson:"title" json:"title"`
	Publisher     string     `bson:"publisher,omitempty" json:"publisher,omitempty"`
	PublishedDate *time.Time `bson:"publishedDate,omitempty" json:"publishedDate,omitempty"`
	Institutional bool       `bson:"institutional" json:"institutional"`
}

type ResearchResponse struct {
	ID             bson.ObjectID `json:"id"`
	PolicyID       bson.ObjectID `json:"policyId"`
	AuthorID       bson.ObjectID `json:"authorId"`
	AuthorUsername string        `json:"authorUsername"`
	AuthorRepTier  string        `json:"authorRepTier"`
	AuthorType     string        `json:"authorType"`
	Title          string        `json:"title"`
	Type           string        `json:"type"`
	Content        string        `json:"content"`
	Sources        []Source      `json:"sources"`
	Upvotes        int           `json:"upvotes"`
	Downvotes      int           `json:"downvotes"`
	Score          int           `json:"score"`
	CitedBy        int           `json:"citedBy"`
	UserVote       int           `json:"userVote"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}
