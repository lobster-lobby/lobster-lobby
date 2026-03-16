package models

import "testing"

func TestTierForScore(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{-5, TierNew},
		{0, TierNew},
		{19, TierNew},
		{20, TierMember},
		{50, TierMember},
		{99, TierMember},
		{100, TierTrusted},
		{150, TierTrusted},
		{199, TierTrusted},
		{200, TierModerator},
		{500, TierModerator},
	}

	for _, tt := range tests {
		got := TierForScore(tt.score)
		if got != tt.expected {
			t.Errorf("TierForScore(%d) = %q, want %q", tt.score, got, tt.expected)
		}
	}
}

func TestPointValues(t *testing.T) {
	expectedActions := []string{
		ActionPolicyCreated,
		ActionCommentPosted,
		ActionResearchSubmitted,
		ActionUpvoteReceived,
		ActionDownvoteReceived,
		ActionEndorsementReceived,
		ActionFlagConfirmed,
		ActionFlagRejected,
		ActionCommentFlaggedConfirm,
	}

	for _, action := range expectedActions {
		if _, ok := PointValues[action]; !ok {
			t.Errorf("missing point value for action %q", action)
		}
	}

	if len(PointValues) != len(expectedActions) {
		t.Errorf("PointValues has %d entries, expected %d", len(PointValues), len(expectedActions))
	}

	// Verify specific values from spec
	checks := map[string]int{
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

	for action, want := range checks {
		if got := PointValues[action]; got != want {
			t.Errorf("PointValues[%q] = %d, want %d", action, got, want)
		}
	}
}
