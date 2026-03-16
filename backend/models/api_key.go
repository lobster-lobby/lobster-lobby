package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type APIKey struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      bson.ObjectID `bson:"userId" json:"userId"`
	Name        string        `bson:"name" json:"name"`
	KeyHash     string        `bson:"keyHash" json:"-"`
	KeyPrefix   string        `bson:"keyPrefix" json:"prefix"`
	Permissions []string      `bson:"permissions" json:"permissions"`
	RateLimit   int           `bson:"rateLimit" json:"rateLimit"`
	LastUsedAt  *time.Time    `bson:"lastUsedAt,omitempty" json:"lastUsedAt"`
	ExpiresAt   *time.Time    `bson:"expiresAt,omitempty" json:"expiresAt"`
	CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`
	Revoked     bool          `bson:"revoked" json:"revoked"`
}
