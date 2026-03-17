package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestCampaignNomination_Validate(t *testing.T) {
	policyID := bson.NewObjectID()
	nominatorID := bson.NewObjectID()

	t.Run("valid nomination", func(t *testing.T) {
		n := &CampaignNomination{
			PolicyID:    policyID,
			NominatedBy: nominatorID,
			Status:      NominationStatusPending,
		}
		if err := n.Validate(); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("defaults status to pending when empty", func(t *testing.T) {
		n := &CampaignNomination{
			PolicyID:    policyID,
			NominatedBy: nominatorID,
		}
		if err := n.Validate(); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if n.Status != NominationStatusPending {
			t.Errorf("expected status %q, got %q", NominationStatusPending, n.Status)
		}
	})

	t.Run("missing policyId", func(t *testing.T) {
		n := &CampaignNomination{
			NominatedBy: nominatorID,
		}
		if err := n.Validate(); err == nil {
			t.Fatal("expected error for missing policyId")
		}
	})

	t.Run("missing nominatedBy", func(t *testing.T) {
		n := &CampaignNomination{
			PolicyID: policyID,
		}
		if err := n.Validate(); err == nil {
			t.Fatal("expected error for missing nominatedBy")
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		n := &CampaignNomination{
			PolicyID:    policyID,
			NominatedBy: nominatorID,
			Status:      NominationStatus("invalid"),
		}
		if err := n.Validate(); err == nil {
			t.Fatal("expected error for invalid status")
		}
	})

	t.Run("approved status is valid", func(t *testing.T) {
		n := &CampaignNomination{
			PolicyID:    policyID,
			NominatedBy: nominatorID,
			Status:      NominationStatusApproved,
		}
		if err := n.Validate(); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("rejected status is valid", func(t *testing.T) {
		n := &CampaignNomination{
			PolicyID:    policyID,
			NominatedBy: nominatorID,
			Status:      NominationStatusRejected,
		}
		if err := n.Validate(); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})
}

func TestNominationConstants(t *testing.T) {
	if EndorsementsRequired <= 0 {
		t.Errorf("EndorsementsRequired must be positive, got %d", EndorsementsRequired)
	}
	if EndorserMinReputation <= 0 {
		t.Errorf("EndorserMinReputation must be positive, got %d", EndorserMinReputation)
	}
}

func TestEndorsementEligibility(t *testing.T) {
	policyID := bson.NewObjectID()
	nominatorID := bson.NewObjectID()

	n := &CampaignNomination{
		PolicyID:    policyID,
		NominatedBy: nominatorID,
		Status:      NominationStatusPending,
		Endorsers:   []NominationEndorsement{},
	}

	// Below threshold
	if len(n.Endorsers) >= EndorsementsRequired {
		t.Error("empty nomination should not meet endorsement threshold")
	}

	// Add endorsers up to the threshold
	for i := 0; i < EndorsementsRequired; i++ {
		n.Endorsers = append(n.Endorsers, NominationEndorsement{
			UserID: bson.NewObjectID(),
		})
	}

	if len(n.Endorsers) < EndorsementsRequired {
		t.Errorf("expected %d endorsers to meet threshold", EndorsementsRequired)
	}
}
