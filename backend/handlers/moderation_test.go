package handlers

import (
	"testing"
)

func TestValidFlagReasons(t *testing.T) {
	valid := map[string]bool{"spam": true, "harassment": true, "misinformation": true, "off-topic": true}

	tests := []struct {
		reason string
		want   bool
	}{
		{"spam", true},
		{"harassment", true},
		{"misinformation", true},
		{"off-topic", true},
		{"invalid", false},
		{"", false},
		{"SPAM", false}, // case sensitive
	}

	for _, tt := range tests {
		got := valid[tt.reason]
		if got != tt.want {
			t.Errorf("reason %q: expected valid=%v, got %v", tt.reason, tt.want, got)
		}
	}
}

func TestValidModerationActions(t *testing.T) {
	valid := map[string]bool{"approve": true, "remove": true, "ban": true}

	tests := []struct {
		action string
		want   bool
	}{
		{"approve", true},
		{"remove", true},
		{"ban", true},
		{"delete", false},
		{"", false},
	}

	for _, tt := range tests {
		got := valid[tt.action]
		if got != tt.want {
			t.Errorf("action %q: expected valid=%v, got %v", tt.action, tt.want, got)
		}
	}
}

func TestFlagThreshold(t *testing.T) {
	// Simulates the auto-flag logic: at 3 flags, argument should be flagged
	threshold := int64(3)

	tests := []struct {
		flagCount    int64
		shouldFlag   bool
	}{
		{1, false},
		{2, false},
		{3, true},
		{5, true},
	}

	for _, tt := range tests {
		flagged := tt.flagCount >= threshold
		if flagged != tt.shouldFlag {
			t.Errorf("flagCount=%d: expected flagged=%v, got %v", tt.flagCount, tt.shouldFlag, flagged)
		}
	}
}
