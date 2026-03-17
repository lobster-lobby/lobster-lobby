package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Reaction struct {
	ID         bson.ObjectID `bson:"_id,omitempty"`
	UserID     bson.ObjectID `bson:"userId"`
	EntityID   bson.ObjectID `bson:"entityId"`
	EntityType string        `bson:"entityType"` // "comment"|"research"
	Value      int           `bson:"value"`      // +1 or -1
	CreatedAt  time.Time     `bson:"createdAt"`
}
