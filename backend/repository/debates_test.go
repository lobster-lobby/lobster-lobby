package repository

import (
	"testing"
)

func TestComputeVoteToggle_NewUpvote(t *testing.T) {
	newVote, delta := ComputeVoteToggle(0, 1)
	if newVote != 1 || delta != 1 {
		t.Errorf("new upvote: expected (1, 1) got (%d, %d)", newVote, delta)
	}
}

func TestComputeVoteToggle_NewDownvote(t *testing.T) {
	newVote, delta := ComputeVoteToggle(0, -1)
	if newVote != -1 || delta != -1 {
		t.Errorf("new downvote: expected (-1, -1) got (%d, %d)", newVote, delta)
	}
}

func TestComputeVoteToggle_RemoveUpvote(t *testing.T) {
	newVote, delta := ComputeVoteToggle(1, 1)
	if newVote != 0 || delta != -1 {
		t.Errorf("remove upvote: expected (0, -1) got (%d, %d)", newVote, delta)
	}
}

func TestComputeVoteToggle_RemoveDownvote(t *testing.T) {
	newVote, delta := ComputeVoteToggle(-1, -1)
	if newVote != 0 || delta != 1 {
		t.Errorf("remove downvote: expected (0, 1) got (%d, %d)", newVote, delta)
	}
}

func TestComputeVoteToggle_ChangeUpToDown(t *testing.T) {
	newVote, delta := ComputeVoteToggle(1, -1)
	if newVote != -1 || delta != -2 {
		t.Errorf("change up->down: expected (-1, -2) got (%d, %d)", newVote, delta)
	}
}

func TestComputeVoteToggle_ChangeDownToUp(t *testing.T) {
	newVote, delta := ComputeVoteToggle(-1, 1)
	if newVote != 1 || delta != 2 {
		t.Errorf("change down->up: expected (1, 2) got (%d, %d)", newVote, delta)
	}
}

// TestDuplicateVotePrevention verifies that repeated same-value votes toggle off.
func TestDuplicateVotePrevention(t *testing.T) {
	// First vote: upvote
	newVote, _ := ComputeVoteToggle(0, 1)
	if newVote != 1 {
		t.Fatalf("expected first upvote to set vote=1, got %d", newVote)
	}

	// Second vote: same upvote (should toggle off / remove)
	newVote, _ = ComputeVoteToggle(1, 1)
	if newVote != 0 {
		t.Errorf("duplicate upvote should toggle to 0, got %d", newVote)
	}

	// Third vote: upvote again (should create new)
	newVote, _ = ComputeVoteToggle(0, 1)
	if newVote != 1 {
		t.Errorf("re-upvote should set vote=1, got %d", newVote)
	}
}

// TestVoteScoreConsistency verifies net score stays consistent through a sequence.
func TestVoteScoreConsistency(t *testing.T) {
	score := 0

	// User A upvotes
	_, delta := ComputeVoteToggle(0, 1)
	score += delta
	if score != 1 {
		t.Fatalf("after upvote: expected score=1, got %d", score)
	}

	// User B downvotes
	_, delta = ComputeVoteToggle(0, -1)
	score += delta
	if score != 0 {
		t.Fatalf("after downvote: expected score=0, got %d", score)
	}

	// User A changes to downvote
	_, delta = ComputeVoteToggle(1, -1)
	score += delta
	if score != -2 {
		t.Fatalf("after change: expected score=-2, got %d", score)
	}

	// User B removes vote
	_, delta = ComputeVoteToggle(-1, -1)
	score += delta
	if score != -1 {
		t.Fatalf("after remove: expected score=-1, got %d", score)
	}
}

// TestSortOrdering_DebateArgumentSortDoc verifies that sort options produce correct BSON sort documents.
func TestSortOrdering_DebateArgumentSortDoc(t *testing.T) {
	repo := &DebateRepository{}

	tests := []struct {
		sort     string
		firstKey string
	}{
		{"newest", "createdAt"},
		{"top", "score"},
		{"unknown", "score"}, // default fallback
	}

	for _, tt := range tests {
		doc := repo.getArgumentSortDoc(tt.sort)
		if len(doc) == 0 {
			t.Errorf("sort=%q: got empty sort document", tt.sort)
			continue
		}
		if doc[0].Key != tt.firstKey {
			t.Errorf("sort=%q: expected first sort key %q, got %q", tt.sort, tt.firstKey, doc[0].Key)
		}
	}
}

// TestSortOrdering_ControversialRanking verifies that controversy scores rank items correctly.
func TestSortOrdering_ControversialRanking(t *testing.T) {
	// Items sorted by expected controversy (most controversial first)
	items := []struct {
		label     string
		upvotes   int
		downvotes int
	}{
		{"even 50/50", 50, 50},      // most controversial: high engagement, even split
		{"even 10/10", 10, 10},      // controversial but less engagement
		{"slight lean 15/10", 15, 10},
		{"one-sided 20/0", 20, 0},   // not controversial at all
		{"no votes", 0, 0},          // no engagement
	}

	prevScore := float64(999)
	for _, item := range items {
		score := ControversyScore(item.upvotes, item.downvotes)
		if score > prevScore {
			t.Errorf("%s: controversy score %f should be <= previous %f", item.label, score, prevScore)
		}
		prevScore = score
	}
}

func TestControversyScore(t *testing.T) {
	tests := []struct {
		name      string
		upvotes   int
		downvotes int
		wantHigh  bool // relative to the "no votes" baseline
	}{
		{"no votes", 0, 0, false},
		{"all upvotes", 10, 0, false},
		{"all downvotes", 0, 10, false},
		{"even split", 10, 10, true},
		{"close split", 11, 9, true},
	}

	noVoteScore := ControversyScore(0, 0)
	allUpScore := ControversyScore(10, 0)
	evenScore := ControversyScore(10, 10)

	if evenScore <= allUpScore {
		t.Errorf("even split should be more controversial than all-upvote: even=%f, allUp=%f", evenScore, allUpScore)
	}

	for _, tt := range tests {
		score := ControversyScore(tt.upvotes, tt.downvotes)
		if tt.wantHigh && score <= noVoteScore {
			t.Errorf("%s: expected controversy score > 0, got %f", tt.name, score)
		}
	}
}

func TestMultipleUsersVoteScore(t *testing.T) {
	score := 0

	// 3 upvotes
	for i := 0; i < 3; i++ {
		_, delta := ComputeVoteToggle(0, 1)
		score += delta
	}
	if score != 3 {
		t.Fatalf("after 3 upvotes: expected 3, got %d", score)
	}

	// 2 downvotes
	for i := 0; i < 2; i++ {
		_, delta := ComputeVoteToggle(0, -1)
		score += delta
	}
	if score != 1 {
		t.Fatalf("after 2 downvotes: expected 1, got %d", score)
	}

	// One user changes upvote to downvote
	_, delta := ComputeVoteToggle(1, -1)
	score += delta
	if score != -1 {
		t.Fatalf("after vote change: expected -1, got %d", score)
	}
}
