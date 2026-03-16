package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	ActionPolicyCreated         = "policy_created"
	ActionCommentPosted         = "comment_posted"
	ActionResearchSubmitted     = "research_submitted"
	ActionUpvoteReceived        = "upvote_received"
	ActionDownvoteReceived      = "downvote_received"
	ActionEndorsementReceived   = "endorsement_received"
	ActionFlagConfirmed         = "flag_confirmed"
	ActionFlagRejected          = "flag_rejected"
	ActionCommentFlaggedConfirm = "comment_flagged_confirmed"
)

var PointValues = map[string]int{
	ActionPolicyCreated:         10,
	ActionCommentPosted:         2,
	ActionResearchSubmitted:     5,
	ActionUpvoteReceived:        1,
	ActionDownvoteReceived:      -1,
	ActionEndorsementReceived:   3,
	ActionFlagConfirmed:         5,
	ActionFlagRejected:          -2,
	ActionCommentFlaggedConfirm: -10,
}

const (
	TierNew       = "new"
	TierMember    = "member"
	TierTrusted   = "trusted"
	TierModerator = "moderator"
)

func TierForScore(score int) string {
	switch {
	case score >= 200:
		return TierModerator
	case score >= 100:
		return TierTrusted
	case score >= 20:
		return TierMember
	default:
		return TierNew
	}
}

type ReputationEvent struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     bson.ObjectID `bson:"userId" json:"userId"`
	Action     string        `bson:"action" json:"action"`
	Points     int           `bson:"points" json:"points"`
	EntityID   string        `bson:"entityId" json:"entityId"`
	EntityType string        `bson:"entityType" json:"entityType"`
	CreatedAt  time.Time     `bson:"createdAt" json:"createdAt"`
}
