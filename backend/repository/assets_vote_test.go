package repository

import (
	"testing"
)

// voteDelta computes (upvoteDelta, downvoteDelta) given an old and new vote value.
// This mirrors the logic inside AssetRepository.SetVote.
func voteDelta(oldValue, newValue int) (upvoteDelta, downvoteDelta int) {
	// Remove old vote effect
	if oldValue == 1 {
		upvoteDelta--
	} else if oldValue == -1 {
		downvoteDelta--
	}

	// Add new vote effect
	if newValue == 1 {
		upvoteDelta++
	} else if newValue == -1 {
		downvoteDelta++
	}

	return
}

// applyDelta simulates the score counters after a vote change.
func applyDelta(upvotes, downvotes, oldValue, newValue int) (int, int, int) {
	ud, dd := voteDelta(oldValue, newValue)
	upvotes += ud
	downvotes += dd
	score := upvotes - downvotes
	return upvotes, downvotes, score
}

func TestVoteDelta_NewUpvote(t *testing.T) {
	// No prior vote → upvote
	up, down := voteDelta(0, 1)
	if up != 1 || down != 0 {
		t.Errorf("expected (1,0) got (%d,%d)", up, down)
	}
}

func TestVoteDelta_NewDownvote(t *testing.T) {
	// No prior vote → downvote
	up, down := voteDelta(0, -1)
	if up != 0 || down != 1 {
		t.Errorf("expected (0,1) got (%d,%d)", up, down)
	}
}

func TestVoteDelta_RemoveUpvote(t *testing.T) {
	// Had upvote → remove (0)
	up, down := voteDelta(1, 0)
	if up != -1 || down != 0 {
		t.Errorf("expected (-1,0) got (%d,%d)", up, down)
	}
}

func TestVoteDelta_RemoveDownvote(t *testing.T) {
	// Had downvote → remove (0)
	up, down := voteDelta(-1, 0)
	if up != 0 || down != -1 {
		t.Errorf("expected (0,-1) got (%d,%d)", up, down)
	}
}

func TestVoteDelta_ChangeUpvoteToDownvote(t *testing.T) {
	// Had upvote → change to downvote
	up, down := voteDelta(1, -1)
	if up != -1 || down != 1 {
		t.Errorf("expected (-1,1) got (%d,%d)", up, down)
	}
}

func TestVoteDelta_ChangeDownvoteToUpvote(t *testing.T) {
	// Had downvote → change to upvote
	up, down := voteDelta(-1, 1)
	if up != 1 || down != -1 {
		t.Errorf("expected (1,-1) got (%d,%d)", up, down)
	}
}

func TestVoteDelta_NoChange(t *testing.T) {
	// No prior, no new
	up, down := voteDelta(0, 0)
	if up != 0 || down != 0 {
		t.Errorf("expected (0,0) got (%d,%d)", up, down)
	}
}

func TestVoteCountConsistency_UpvoteThenRemove(t *testing.T) {
	upvotes, downvotes := 0, 0
	upvotes, downvotes, score := applyDelta(upvotes, downvotes, 0, 1) // upvote
	if upvotes != 1 || downvotes != 0 || score != 1 {
		t.Errorf("after upvote: want (1,0,1) got (%d,%d,%d)", upvotes, downvotes, score)
	}
	upvotes, downvotes, score = applyDelta(upvotes, downvotes, 1, 0) // remove
	if upvotes != 0 || downvotes != 0 || score != 0 {
		t.Errorf("after remove: want (0,0,0) got (%d,%d,%d)", upvotes, downvotes, score)
	}
}

func TestVoteCountConsistency_DownvoteThenRemove(t *testing.T) {
	upvotes, downvotes := 0, 0
	upvotes, downvotes, score := applyDelta(upvotes, downvotes, 0, -1) // downvote
	if upvotes != 0 || downvotes != 1 || score != -1 {
		t.Errorf("after downvote: want (0,1,-1) got (%d,%d,%d)", upvotes, downvotes, score)
	}
	upvotes, downvotes, score = applyDelta(upvotes, downvotes, -1, 0) // remove
	if upvotes != 0 || downvotes != 0 || score != 0 {
		t.Errorf("after remove: want (0,0,0) got (%d,%d,%d)", upvotes, downvotes, score)
	}
}

func TestVoteCountConsistency_ChangeVote(t *testing.T) {
	// Start with upvote, change to downvote
	upvotes, downvotes := 0, 0
	upvotes, downvotes, score := applyDelta(upvotes, downvotes, 0, 1)   // upvote
	upvotes, downvotes, score = applyDelta(upvotes, downvotes, 1, -1)   // change to downvote
	if upvotes != 0 || downvotes != 1 || score != -1 {
		t.Errorf("after change upvote→downvote: want (0,1,-1) got (%d,%d,%d)", upvotes, downvotes, score)
	}
	upvotes, downvotes, score = applyDelta(upvotes, downvotes, -1, 1)   // change back to upvote
	if upvotes != 1 || downvotes != 0 || score != 1 {
		t.Errorf("after change downvote→upvote: want (1,0,1) got (%d,%d,%d)", upvotes, downvotes, score)
	}
}

func TestVoteCountConsistency_MultipleUsers(t *testing.T) {
	// Simulate 3 users voting independently (totals accumulate)
	upvotes, downvotes := 0, 0
	// User 1: upvote
	ud, dd := voteDelta(0, 1)
	upvotes += ud; downvotes += dd
	// User 2: downvote
	ud, dd = voteDelta(0, -1)
	upvotes += ud; downvotes += dd
	// User 3: upvote
	ud, dd = voteDelta(0, 1)
	upvotes += ud; downvotes += dd

	score := upvotes - downvotes
	if upvotes != 2 || downvotes != 1 || score != 1 {
		t.Errorf("multi-user: want (2,1,1) got (%d,%d,%d)", upvotes, downvotes, score)
	}

	// User 2 changes downvote to upvote
	ud, dd = voteDelta(-1, 1)
	upvotes += ud; downvotes += dd
	score = upvotes - downvotes
	if upvotes != 3 || downvotes != 0 || score != 3 {
		t.Errorf("after user2 changes: want (3,0,3) got (%d,%d,%d)", upvotes, downvotes, score)
	}
}
