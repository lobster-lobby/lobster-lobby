package models

import (
	"testing"
	"time"
)

func validCampaign() *Campaign {
	return &Campaign{
		Title:       "Test Campaign Title",
		Objective:   "This is a valid objective with enough chars",
		Target:      "Target audience here",
		Description: "This is a valid description that is long enough to pass validation checks.",
		Status:      CampaignStatusActive,
	}
}

func TestCampaign_Validate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Campaign)
		wantErr bool
	}{
		{
			name:    "valid campaign",
			modify:  func(c *Campaign) {},
			wantErr: false,
		},
		{
			name:    "title too short",
			modify:  func(c *Campaign) { c.Title = "abc" },
			wantErr: true,
		},
		{
			name:    "title too long",
			modify:  func(c *Campaign) { c.Title = string(make([]byte, 201)) },
			wantErr: true,
		},
		{
			name:    "objective too short",
			modify:  func(c *Campaign) { c.Objective = "short" },
			wantErr: true,
		},
		{
			name:    "target too short",
			modify:  func(c *Campaign) { c.Target = "ab" },
			wantErr: true,
		},
		{
			name:    "description too short",
			modify:  func(c *Campaign) { c.Description = "too short" },
			wantErr: true,
		},
		{
			name:    "invalid status",
			modify:  func(c *Campaign) { c.Status = "invalid" },
			wantErr: true,
		},
		{
			name:    "empty status defaults to active",
			modify:  func(c *Campaign) { c.Status = "" },
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validCampaign()
			tt.modify(c)
			err := c.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"  Leading and trailing  ", "leading-and-trailing"},
		{"Multiple   Spaces", "multiple-spaces"},
		{"Special!@#$%Chars", "special-chars"},
		{"Already-Slugged", "already-slugged"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			got := GenerateSlug(tt.title)
			if got != tt.expected {
				t.Errorf("GenerateSlug(%q) = %q, want %q", tt.title, got, tt.expected)
			}
		})
	}
}

func TestGenerateSlug_Truncation(t *testing.T) {
	// Build a 110-char title that would produce trailing dashes at boundary
	// "aaaa...aaaa-bbb" where the cut at 100 would land in the middle of a word
	title := "a very long campaign title that goes well beyond the one hundred character limit for slug generation here"
	slug := GenerateSlug(title)
	if len(slug) > 100 {
		t.Errorf("slug length %d exceeds 100", len(slug))
	}
	if len(slug) > 0 && slug[len(slug)-1] == '-' {
		t.Errorf("slug has trailing dash: %q", slug)
	}
}

func TestCampaign_CalculateTrendingScore(t *testing.T) {
	t.Run("zero activity gives zero score", func(t *testing.T) {
		c := &Campaign{
			CreatedAt: time.Now().Add(-2 * time.Hour),
		}
		score := c.CalculateTrendingScore()
		if score != 0 {
			t.Errorf("expected 0 score, got %f", score)
		}
	})

	t.Run("more activity means higher score for same age", func(t *testing.T) {
		created := time.Now().Add(-1 * time.Hour)
		low := &Campaign{
			CreatedAt: created,
			Metrics:   CampaignMetrics{TotalShares: 1, UniqueParticipants: 1, CommentCount: 1},
		}
		high := &Campaign{
			CreatedAt: created,
			Metrics:   CampaignMetrics{TotalShares: 100, UniqueParticipants: 100, CommentCount: 100},
		}
		if high.CalculateTrendingScore() <= low.CalculateTrendingScore() {
			t.Error("expected higher activity to yield higher trending score")
		}
	})

	t.Run("older campaign scores lower than newer with same activity", func(t *testing.T) {
		metrics := CampaignMetrics{TotalShares: 10, UniqueParticipants: 10, CommentCount: 10}
		newer := &Campaign{CreatedAt: time.Now().Add(-1 * time.Hour), Metrics: metrics}
		older := &Campaign{CreatedAt: time.Now().Add(-24 * time.Hour), Metrics: metrics}
		if newer.CalculateTrendingScore() <= older.CalculateTrendingScore() {
			t.Error("expected newer campaign to score higher than older one")
		}
	})
}
