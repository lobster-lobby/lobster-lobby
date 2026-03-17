package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CrossReference struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	SourceType string        `bson:"sourceType" json:"sourceType"` // "research"|"debate"|"policy"
	SourceID   bson.ObjectID `bson:"sourceId" json:"sourceId"`
	TargetType string        `bson:"targetType" json:"targetType"` // "research"|"debate"|"policy"
	TargetID   bson.ObjectID `bson:"targetId" json:"targetId"`
	CreatedBy  bson.ObjectID `bson:"createdBy" json:"createdBy"`
	CreatedAt  time.Time     `bson:"createdAt" json:"createdAt"`
}

type CrossReferenceResponse struct {
	ID         bson.ObjectID `json:"id"`
	SourceType string        `json:"sourceType"`
	SourceID   bson.ObjectID `json:"sourceId"`
	SourceTitle string       `json:"sourceTitle"`
	TargetType string        `json:"targetType"`
	TargetID   bson.ObjectID `json:"targetId"`
	TargetTitle string       `json:"targetTitle"`
	CreatedBy  bson.ObjectID `json:"createdBy"`
	CreatedAt  time.Time     `json:"createdAt"`
}
