package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Representative represents an elected official stored in the database
type Representative struct {
	ID          bson.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name        string            `bson:"name" json:"name"`
	Title       string            `bson:"title" json:"title"`
	Party       string            `bson:"party" json:"party"`
	State       string            `bson:"state" json:"state"`
	District    string            `bson:"district" json:"district"`
	PhotoURL    string            `bson:"photoUrl,omitempty" json:"photoUrl,omitempty"`
	Phone       string            `bson:"phone,omitempty" json:"phone,omitempty"`
	Email       string            `bson:"email,omitempty" json:"email,omitempty"`
	Website     string            `bson:"website,omitempty" json:"website,omitempty"`
	SocialMedia map[string]string `bson:"socialMedia,omitempty" json:"socialMedia,omitempty"`
	Chamber     string            `bson:"chamber" json:"chamber"` // senate, house, governor, local
	Level       string            `bson:"level" json:"level"`     // federal, state, local
	ExternalIDs ExternalIDs       `bson:"externalIds,omitempty" json:"externalIds,omitempty"`
	CreatedAt   time.Time         `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time         `bson:"updatedAt" json:"updatedAt"`
}

// ExternalIDs holds various external identifiers for a representative
type ExternalIDs struct {
	BioguideID string `bson:"bioguideId,omitempty" json:"bioguideId,omitempty"`
	GovtrackID string `bson:"govtrackId,omitempty" json:"govtrackId,omitempty"`
}

// CivicOfficial represents an official returned from the Google Civic API
// This is used for address-based lookups and is NOT stored in the database
type CivicOfficial struct {
	Name        string            `json:"name"`
	Title       string            `json:"title"`
	Party       string            `json:"party"`
	Phone       string            `json:"phone,omitempty"`
	Email       string            `json:"email,omitempty"`
	PhotoURL    string            `json:"photoUrl,omitempty"`
	URLs        []string          `json:"urls,omitempty"`
	SocialMedia map[string]string `json:"socialMedia,omitempty"`
}
