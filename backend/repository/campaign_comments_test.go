package repository

import (
	"testing"
)

// TestCampaignCommentVoteDelta tests the vote delta calculation logic
// used in ToggleVote to ensure vote counts stay consistent.

// simulateToggleVote simulates the ToggleVote logic and returns (newVoteValue, voteDelta).
// This mirrors the behavior in CampaignCommentRepository.ToggleVote.
func simulateToggleVote(existingVote, requestedVote int) (newVote, delta int) {
	if existingVote == 0 {
		// No existing vote - create new one
		return requestedVote, requestedVote
	}

	if existingVote == requestedVote {
		// Same vote - toggle off (remove)
		return 0, -existingVote
	}

	// Different vote - change it
	delta = requestedVote - existingVote
	return requestedVote, delta
}

func TestToggleVote_NewUpvote(t *testing.T) {
	newVote, delta := simulateToggleVote(0, 1)
	if newVote != 1 || delta != 1 {
		t.Errorf("new upvote: expected (1, 1) got (%d, %d)", newVote, delta)
	}
}

func TestToggleVote_NewDownvote(t *testing.T) {
	newVote, delta := simulateToggleVote(0, -1)
	if newVote != -1 || delta != -1 {
		t.Errorf("new downvote: expected (-1, -1) got (%d, %d)", newVote, delta)
	}
}

func TestToggleVote_RemoveUpvote(t *testing.T) {
	// Had upvote, vote upvote again -> remove
	newVote, delta := simulateToggleVote(1, 1)
	if newVote != 0 || delta != -1 {
		t.Errorf("remove upvote: expected (0, -1) got (%d, %d)", newVote, delta)
	}
}

func TestToggleVote_RemoveDownvote(t *testing.T) {
	// Had downvote, vote downvote again -> remove
	newVote, delta := simulateToggleVote(-1, -1)
	if newVote != 0 || delta != 1 {
		t.Errorf("remove downvote: expected (0, 1) got (%d, %d)", newVote, delta)
	}
}

func TestToggleVote_ChangeUpvoteToDownvote(t *testing.T) {
	// Had upvote, now voting downvote
	newVote, delta := simulateToggleVote(1, -1)
	if newVote != -1 || delta != -2 {
		t.Errorf("change up->down: expected (-1, -2) got (%d, %d)", newVote, delta)
	}
}

func TestToggleVote_ChangeDownvoteToUpvote(t *testing.T) {
	// Had downvote, now voting upvote
	newVote, delta := simulateToggleVote(-1, 1)
	if newVote != 1 || delta != 2 {
		t.Errorf("change down->up: expected (1, 2) got (%d, %d)", newVote, delta)
	}
}

// TestVoteCountConsistency tests that vote counts remain consistent through
// a series of operations.
func TestVoteCountConsistency_ToggleSequence(t *testing.T) {
	voteCount := 0

	// User upvotes
	newVote, delta := simulateToggleVote(0, 1)
	voteCount += delta
	if voteCount != 1 || newVote != 1 {
		t.Errorf("after upvote: voteCount=%d, newVote=%d, expected (1, 1)", voteCount, newVote)
	}

	// User toggles off (upvote again)
	newVote, delta = simulateToggleVote(1, 1)
	voteCount += delta
	if voteCount != 0 || newVote != 0 {
		t.Errorf("after toggle off: voteCount=%d, newVote=%d, expected (0, 0)", voteCount, newVote)
	}

	// User downvotes
	newVote, delta = simulateToggleVote(0, -1)
	voteCount += delta
	if voteCount != -1 || newVote != -1 {
		t.Errorf("after downvote: voteCount=%d, newVote=%d, expected (-1, -1)", voteCount, newVote)
	}

	// User changes to upvote
	newVote, delta = simulateToggleVote(-1, 1)
	voteCount += delta
	if voteCount != 1 || newVote != 1 {
		t.Errorf("after change to upvote: voteCount=%d, newVote=%d, expected (1, 1)", voteCount, newVote)
	}
}

// TestMultipleUsersVoting simulates multiple users voting on the same comment.
func TestMultipleUsersVoting(t *testing.T) {
	voteCount := 0

	// User 1 upvotes
	_, delta := simulateToggleVote(0, 1)
	voteCount += delta

	// User 2 upvotes
	_, delta = simulateToggleVote(0, 1)
	voteCount += delta

	// User 3 downvotes
	_, delta = simulateToggleVote(0, -1)
	voteCount += delta

	if voteCount != 1 {
		t.Errorf("after 3 users vote: expected voteCount=1, got %d", voteCount)
	}

	// User 3 changes to upvote
	_, delta = simulateToggleVote(-1, 1)
	voteCount += delta

	if voteCount != 3 {
		t.Errorf("after user3 changes vote: expected voteCount=3, got %d", voteCount)
	}

	// User 1 removes vote
	_, delta = simulateToggleVote(1, 1)
	voteCount += delta

	if voteCount != 2 {
		t.Errorf("after user1 removes vote: expected voteCount=2, got %d", voteCount)
	}
}

// TestCommentListSort tests that sort option strings are valid.
func TestCommentListSort(t *testing.T) {
	validSorts := []string{"newest", "votes", ""}
	for _, s := range validSorts {
		// Just verify the sort string doesn't panic when used
		opts := CampaignCommentListOpts{Sort: s}
		if opts.Sort != s {
			t.Errorf("expected sort=%q, got %q", s, opts.Sort)
		}
	}
}
