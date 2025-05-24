package executor

import (
	"testing"
	"time"

	"github.com/zinrai/x-scheduler/internal/config"
)

func TestShouldPostNow(t *testing.T) {
	baseTime := time.Date(2024, 6, 1, 12, 30, 0, 0, time.UTC)

	tests := []struct {
		name        string
		scheduledAt time.Time
		currentTime time.Time
		want        bool
	}{
		{
			name:        "exact time match should return true",
			scheduledAt: baseTime,
			currentTime: baseTime,
			want:        true,
		},
		{
			name:        "within 1 minute window should return true",
			scheduledAt: baseTime,
			currentTime: baseTime.Add(30 * time.Second),
			want:        true,
		},
		{
			name:        "exactly 1 minute difference should return true",
			scheduledAt: baseTime,
			currentTime: baseTime.Add(1 * time.Minute),
			want:        true,
		},
		{
			name:        "over 1 minute difference should return false",
			scheduledAt: baseTime,
			currentTime: baseTime.Add(2 * time.Minute),
			want:        false,
		},
		{
			name:        "1 minute before should return true",
			scheduledAt: baseTime,
			currentTime: baseTime.Add(-1 * time.Minute),
			want:        true,
		},
		{
			name:        "over 1 minute before should return false",
			scheduledAt: baseTime,
			currentTime: baseTime.Add(-2 * time.Minute),
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldPostNow(tt.scheduledAt, tt.currentTime)
			if got != tt.want {
				t.Errorf("ShouldPostNow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindMatchingPosts(t *testing.T) {
	currentTime := time.Date(2024, 6, 1, 12, 30, 0, 0, time.UTC)
	matchingTime := currentTime
	nonMatchingTime := currentTime.Add(5 * time.Minute)

	posts := []config.Post{
		{
			Content:     "Matching regular post",
			ScheduledAt: matchingTime,
			Enabled:     true,
		},
		{
			Content:     "Non-matching regular post",
			ScheduledAt: nonMatchingTime,
			Enabled:     true,
		},
		{
			Content:     "Test post should always match",
			ScheduledAt: nonMatchingTime,
			Enabled:     true,
			Test:        true,
		},
	}

	matches := FindMatchingPosts(posts, currentTime)

	if len(matches) != 2 {
		t.Errorf("FindMatchingPosts() returned %d posts, want 2", len(matches))
	}

	// Check that we got the right posts
	expectedContents := map[string]bool{
		"Matching regular post":         false,
		"Test post should always match": false,
	}

	for _, match := range matches {
		if _, exists := expectedContents[match.Content]; exists {
			expectedContents[match.Content] = true
		} else {
			t.Errorf("FindMatchingPosts() returned unexpected post: %s", match.Content)
		}
	}

	for content, found := range expectedContents {
		if !found {
			t.Errorf("FindMatchingPosts() missing expected post: %s", content)
		}
	}
}
