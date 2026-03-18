package models

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// VoteType represents the type of vote cast
type VoteType string

const (
	VoteYea     VoteType = "yea"
	VoteNay     VoteType = "nay"
	VoteAbstain VoteType = "abstain"
	VoteAbsent  VoteType = "absent"
)

// VotingRecord represents a single vote cast by a representative on a policy
type VotingRecord struct {
	ID               bson.ObjectID `bson:"_id,omitempty" json:"id"`
	RepresentativeID bson.ObjectID `bson:"representativeId" json:"representativeId"`
	PolicyID         bson.ObjectID `bson:"policyId" json:"policyId"`
	Vote             VoteType      `bson:"vote" json:"vote"`
	Date             time.Time     `bson:"date" json:"date"`
	Session          string        `bson:"session" json:"session"`
	Notes            string        `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt        time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt        time.Time     `bson:"updatedAt" json:"updatedAt"`
}

// VotingSummary provides aggregate voting statistics for a representative
type VotingSummary struct {
	TotalVotes     int     `json:"totalVotes"`
	YeaCount       int     `json:"yeaCount"`
	NayCount       int     `json:"nayCount"`
	AbstainCount   int     `json:"abstainCount"`
	AbsentCount    int     `json:"absentCount"`
	YeaPercent     float64 `json:"yeaPercent"`
	NayPercent     float64 `json:"nayPercent"`
	AbstainPercent float64 `json:"abstainPercent"`
}

// Validate checks that required fields are present and valid
func (v *VotingRecord) Validate() error {
	if v.RepresentativeID.IsZero() {
		return errors.New("representativeId is required")
	}
	if v.PolicyID.IsZero() {
		return errors.New("policyId is required")
	}
	switch v.Vote {
	case VoteYea, VoteNay, VoteAbstain, VoteAbsent:
		// valid
	default:
		return errors.New("vote must be one of: yea, nay, abstain, absent")
	}
	if v.Date.IsZero() {
		return errors.New("date is required")
	}
	if v.Session == "" {
		return errors.New("session is required")
	}
	return nil
}
