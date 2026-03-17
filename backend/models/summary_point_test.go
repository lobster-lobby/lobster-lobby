package models

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func makeEndorsement(position, repTier string, verified bool) Endorsement {
	return Endorsement{
		UserID:    bson.NewObjectID(),
		Position:  position,
		RepTier:   repTier,
		Verified:  verified,
		CreatedAt: time.Now(),
	}
}

func TestCalculateBridgingScore_Empty(t *testing.T) {
	score := CalculateBridgingScore(PositionSupport, nil)
	if score != 0 {
		t.Errorf("expected 0 for empty endorsements, got %f", score)
	}
}

func TestCalculateBridgingScore_SamePosition(t *testing.T) {
	// Same-position endorsement: base=1, rep=1.0 (new), verified=false → 1.0
	e := makeEndorsement(PositionSupport, TierNew, false)
	score := CalculateBridgingScore(PositionSupport, []Endorsement{e})
	if score != 1.0 {
		t.Errorf("expected 1.0, got %f", score)
	}
}

func TestCalculateBridgingScore_CrossPosition(t *testing.T) {
	// Cross-position endorsement: base=3, rep=1.0 (new), verified=false → 3.0
	e := makeEndorsement(PositionOppose, TierNew, false)
	score := CalculateBridgingScore(PositionSupport, []Endorsement{e})
	if score != 3.0 {
		t.Errorf("expected 3.0 for cross-position, got %f", score)
	}
}

func TestCalculateBridgingScore_ConsensusEndorserIsNotCross(t *testing.T) {
	// Consensus endorser on a support point: not cross-position → base=1
	e := makeEndorsement(PositionConsensus, TierNew, false)
	score := CalculateBridgingScore(PositionSupport, []Endorsement{e})
	if score != 1.0 {
		t.Errorf("expected 1.0 for consensus endorser (not cross), got %f", score)
	}
}

func TestCalculateBridgingScore_RepMultipliers(t *testing.T) {
	tests := []struct {
		repTier  string
		expected float64
	}{
		{TierNew, 1.0},
		{TierMember, 1.1},
		{TierTrusted, 1.2},
		{TierModerator, 1.5},
	}
	for _, tt := range tests {
		e := makeEndorsement(PositionSupport, tt.repTier, false)
		score := CalculateBridgingScore(PositionSupport, []Endorsement{e})
		if score != tt.expected {
			t.Errorf("rep tier %s: expected %f, got %f", tt.repTier, tt.expected, score)
		}
	}
}

func TestCalculateBridgingScore_VerifiedMultiplier(t *testing.T) {
	// Same-position, new tier, verified: 1.0 * 1.0 * 1.5 = 1.5
	e := makeEndorsement(PositionSupport, TierNew, true)
	score := CalculateBridgingScore(PositionSupport, []Endorsement{e})
	if score != 1.5 {
		t.Errorf("expected 1.5 for verified endorser, got %f", score)
	}
}

func TestCalculateBridgingScore_CrossPositionVerifiedModerator(t *testing.T) {
	// Cross-position + moderator + verified: 3.0 * 1.5 * 1.5 = 6.75
	e := makeEndorsement(PositionOppose, TierModerator, true)
	score := CalculateBridgingScore(PositionSupport, []Endorsement{e})
	expected := 3.0 * 1.5 * 1.5
	if score != expected {
		t.Errorf("expected %f, got %f", expected, score)
	}
}

func TestCalculateBridgingScore_MultipleEndorsements(t *testing.T) {
	// same-pos new unverified (1.0) + cross-pos trusted unverified (3*1.2=3.6) = 4.6
	e1 := makeEndorsement(PositionSupport, TierNew, false)
	e2 := makeEndorsement(PositionOppose, TierTrusted, false)
	score := CalculateBridgingScore(PositionSupport, []Endorsement{e1, e2})
	expected := 1.0 + 3.0*1.2
	if score != expected {
		t.Errorf("expected %f, got %f", expected, score)
	}
}

func TestBridgingVisibilityThreshold(t *testing.T) {
	// Two cross-position moderator verified endorsements should exceed threshold (5.0)
	// 3 * 1.5 * 1.5 = 6.75 each; one is enough
	e := makeEndorsement(PositionOppose, TierModerator, true)
	score := CalculateBridgingScore(PositionSupport, []Endorsement{e})
	if score < BridgingVisibilityThreshold {
		t.Errorf("expected score %f to exceed visibility threshold %f", score, BridgingVisibilityThreshold)
	}
}
