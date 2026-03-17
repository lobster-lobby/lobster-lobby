package models

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type NominationStatus string

const (
	NominationStatusPending  NominationStatus = "pending"
	NominationStatusApproved NominationStatus = "approved"
	NominationStatusRejected NominationStatus = "rejected"
)

type NominationEndorsement struct {
	UserID    bson.ObjectID `bson:"userId" json:"userId"`
	CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
}

type CampaignNomination struct {
	ID          bson.ObjectID           `bson:"_id,omitempty" json:"id"`
	PolicyID    bson.ObjectID           `bson:"policyId" json:"policyId"`
	NominatedBy bson.ObjectID           `bson:"nominatedBy" json:"nominatedBy"`
	Endorsers   []NominationEndorsement `bson:"endorsers" json:"endorsers"`
	Status      NominationStatus        `bson:"status" json:"status"`
	CreatedAt   time.Time               `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time               `bson:"updatedAt" json:"updatedAt"`
}

func (n *CampaignNomination) Validate() error {
	if n.PolicyID.IsZero() {
		return errors.New("policyId is required")
	}
	if n.NominatedBy.IsZero() {
		return errors.New("nominatedBy is required")
	}
	if n.Status == "" {
		n.Status = NominationStatusPending
	} else if n.Status != NominationStatusPending && n.Status != NominationStatusApproved && n.Status != NominationStatusRejected {
		return errors.New("status must be one of: pending, approved, rejected")
	}
	return nil
}

const EndorsementsRequired = 5
const EndorserMinReputation = 50
