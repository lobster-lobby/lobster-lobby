package repository

import (
	"math"
	"testing"

	"github.com/lobster-lobby/lobster-lobby/models"
)

func TestComputeResearchVoteToggle_NewUpvote(t *testing.T) {
	newVote, delta := ComputeResearchVoteToggle(0, 1)
	if newVote != 1 || delta != 1 {
		t.Errorf("new upvote: expected (1, 1) got (%d, %d)", newVote, delta)
	}
}

func TestComputeResearchVoteToggle_NewDownvote(t *testing.T) {
	newVote, delta := ComputeResearchVoteToggle(0, -1)
	if newVote != -1 || delta != -1 {
		t.Errorf("new downvote: expected (-1, -1) got (%d, %d)", newVote, delta)
	}
}

func TestComputeResearchVoteToggle_RemoveUpvote(t *testing.T) {
	newVote, delta := ComputeResearchVoteToggle(1, 1)
	if newVote != 0 || delta != -1 {
		t.Errorf("toggle off upvote: expected (0, -1) got (%d, %d)", newVote, delta)
	}
}

func TestComputeResearchVoteToggle_RemoveDownvote(t *testing.T) {
	newVote, delta := ComputeResearchVoteToggle(-1, -1)
	if newVote != 0 || delta != 1 {
		t.Errorf("toggle off downvote: expected (0, 1) got (%d, %d)", newVote, delta)
	}
}

func TestComputeResearchVoteToggle_ChangeUpToDown(t *testing.T) {
	newVote, delta := ComputeResearchVoteToggle(1, -1)
	if newVote != -1 || delta != -2 {
		t.Errorf("change up->down: expected (-1, -2) got (%d, %d)", newVote, delta)
	}
}

func TestComputeResearchVoteToggle_ChangeDownToUp(t *testing.T) {
	newVote, delta := ComputeResearchVoteToggle(-1, 1)
	if newVote != 1 || delta != 2 {
		t.Errorf("change down->up: expected (1, 2) got (%d, %d)", newVote, delta)
	}
}

// TestResearchDuplicateVotePrevention verifies that repeated same-value votes toggle off.
func TestResearchDuplicateVotePrevention(t *testing.T) {
	// First vote: upvote
	newVote, _ := ComputeResearchVoteToggle(0, 1)
	if newVote != 1 {
		t.Fatalf("expected first upvote to set vote=1, got %d", newVote)
	}

	// Second vote: same upvote (should toggle off)
	newVote, _ = ComputeResearchVoteToggle(1, 1)
	if newVote != 0 {
		t.Errorf("duplicate upvote should toggle to 0, got %d", newVote)
	}

	// Third vote: upvote again (should create new)
	newVote, _ = ComputeResearchVoteToggle(0, 1)
	if newVote != 1 {
		t.Errorf("re-upvote should set vote=1, got %d", newVote)
	}
}

// TestResearchSortOrdering verifies that sort options produce correct BSON sort documents.
func TestResearchSortOrdering(t *testing.T) {
	repo := &ResearchRepository{}

	tests := []struct {
		sort     string
		firstKey string
	}{
		{"newest", "createdAt"},
		{"top", "score"},
		{"quality", "qualityScore"},
		{"most_cited", "citedBy"},
		{"unknown", "createdAt"}, // default fallback
	}

	for _, tt := range tests {
		doc := repo.getSortDoc(tt.sort)
		if len(doc) == 0 {
			t.Errorf("sort=%q: got empty sort document", tt.sort)
			continue
		}
		if doc[0].Key != tt.firstKey {
			t.Errorf("sort=%q: expected first sort key %q, got %q", tt.sort, tt.firstKey, doc[0].Key)
		}
	}
}

func TestComputeQualityScore_NoVotes_NoSources(t *testing.T) {
	score := ComputeQualityScore(0, 0, nil)
	if score != 0 {
		t.Errorf("expected 0 for no votes/no sources, got %f", score)
	}
}

func TestComputeQualityScore_NoVotes_WithSources(t *testing.T) {
	sources := []models.Source{
		{URL: "https://example.com", Title: "test", Institutional: false},
	}
	score := ComputeQualityScore(0, 0, sources)
	// 0.3 * 50 = 15
	if math.Abs(score-15) > 0.01 {
		t.Errorf("expected ~15 for non-institutional source with no votes, got %f", score)
	}
}

func TestComputeQualityScore_NoVotes_InstitutionalSources(t *testing.T) {
	sources := []models.Source{
		{URL: "https://cbo.gov", Title: "CBO report", Institutional: true},
	}
	score := ComputeQualityScore(0, 0, sources)
	// (0.3 + 0.7) * 50 = 50
	if math.Abs(score-50) > 0.01 {
		t.Errorf("expected ~50 for institutional source with no votes, got %f", score)
	}
}

func TestComputeQualityScore_AllUpvotes(t *testing.T) {
	sources := []models.Source{
		{URL: "https://cbo.gov", Title: "CBO report", Institutional: true},
	}
	score := ComputeQualityScore(10, 0, sources)
	// voteRatio=1.0, credibility=1.0, engagement=log2(11)≈3.459
	// (1.0*0.6 + 1.0*0.4) * 20 * 3.459 ≈ 69.19
	if score < 60 || score > 80 {
		t.Errorf("expected score in range 60-80 for all upvotes+institutional, got %f", score)
	}
}

func TestComputeQualityScore_MixedVotes(t *testing.T) {
	sources := []models.Source{
		{URL: "https://example.com", Title: "blog", Institutional: false},
	}
	score := ComputeQualityScore(5, 5, sources)
	// voteRatio=0.5, credibility=0.3, engagement=log2(11)≈3.459
	// (0.5*0.6 + 0.3*0.4) * 20 * 3.459 ≈ (0.42) * 69.19 ≈ 29.06
	if score < 20 || score > 40 {
		t.Errorf("expected score in range 20-40 for mixed votes, got %f", score)
	}
}

func TestComputeQualityScore_HigherIsBetter(t *testing.T) {
	goodSources := []models.Source{
		{URL: "https://cbo.gov", Title: "CBO", Institutional: true},
	}
	badSources := []models.Source{
		{URL: "https://example.com", Title: "blog", Institutional: false},
	}

	goodScore := ComputeQualityScore(20, 2, goodSources)
	badScore := ComputeQualityScore(5, 10, badSources)

	if goodScore <= badScore {
		t.Errorf("good research (%f) should score higher than bad research (%f)", goodScore, badScore)
	}
}
