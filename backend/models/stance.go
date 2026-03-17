package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Stance struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	UserID    bson.ObjectID `bson:"userId"`
	PolicyID  bson.ObjectID `bson:"policyId"`
	Position  string        `bson:"position"` // "support"|"oppose"|"neutral"
	UpdatedAt time.Time     `bson:"updatedAt"`
}
