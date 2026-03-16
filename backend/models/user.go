package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID                bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	Username          string          `bson:"username" json:"username"`
	Email             string          `bson:"email,omitempty" json:"email,omitempty"`
	PasswordHash      string          `bson:"passwordHash" json:"-"`
	Type              string          `bson:"type" json:"type"`
	Verified          bool            `bson:"verified" json:"verified"`
	VerificationLevel string          `bson:"verificationLevel" json:"verificationLevel"`
	DisplayName       string          `bson:"displayName" json:"displayName"`
	Bio               string          `bson:"bio" json:"bio"`
	Reputation        ReputationScore `bson:"reputation" json:"reputation"`
	Bookmarks         []bson.ObjectID `bson:"bookmarks" json:"bookmarks"`
	District          *District       `bson:"district,omitempty" json:"district,omitempty"`
	CreatedAt         time.Time       `bson:"createdAt" json:"createdAt"`
	UpdatedAt         time.Time       `bson:"updatedAt" json:"updatedAt"`
	LastLoginAt       *time.Time      `bson:"lastLoginAt,omitempty" json:"lastLoginAt,omitempty"`
}

type ReputationScore struct {
	Score         int    `bson:"score" json:"score"`
	Contributions int    `bson:"contributions" json:"contributions"`
	Tier          string `bson:"tier" json:"tier"`
}

type District struct {
	State                 string `bson:"state" json:"state"`
	CongressionalDistrict string `bson:"congressionalDistrict" json:"congressionalDistrict"`
}
