package models

import (
	"errors"
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
	Bio         string            `bson:"bio,omitempty" json:"bio,omitempty"`
	ContactInfo ContactInfo       `bson:"contactInfo,omitempty" json:"contactInfo,omitempty"`
	ExternalIDs ExternalIDs       `bson:"externalIds,omitempty" json:"externalIds,omitempty"`
	CreatedAt   time.Time         `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time         `bson:"updatedAt" json:"updatedAt"`
}

// ContactInfo holds structured contact details for a representative
type ContactInfo struct {
	Phone   string `bson:"phone,omitempty" json:"phone,omitempty"`
	Email   string `bson:"email,omitempty" json:"email,omitempty"`
	Website string `bson:"website,omitempty" json:"website,omitempty"`
	Office  string `bson:"office,omitempty" json:"office,omitempty"`
}

// Validate checks that required fields are present and valid
func (r *Representative) Validate() error {
	if len(r.Name) < 2 || len(r.Name) > 200 {
		return errors.New("name must be between 2 and 200 characters")
	}
	if r.Party == "" {
		return errors.New("party is required")
	}
	if r.State == "" {
		return errors.New("state is required")
	}
	if r.Chamber == "" {
		return errors.New("chamber is required")
	}
	validChambers := map[string]bool{"senate": true, "house": true, "governor": true, "local": true}
	if !validChambers[r.Chamber] {
		return errors.New("chamber must be one of: senate, house, governor, local")
	}
	return nil
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
