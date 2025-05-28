package executor

import (
	"testing"
	"time"

	"github.com/zinrai/x-scheduler/internal/config"
)

func TestIsTodayAndFuture(t *testing.T) {
	// Use a fixed current time for testing
	currentTime := time.Date(2024, 6, 1, 12, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		postTime time.Time
		want     bool
	}{
		{
			name:     "post scheduled for later today should return true",
			postTime: time.Date(2024, 6, 1, 15, 0, 0, 0, time.UTC),
			want:     true,
		},
		{
			name:     "post scheduled for earlier today should return false",
			postTime: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC),
			want:     false,
		},
		{
			name:     "post scheduled for tomorrow should return false",
			postTime: time.Date(2024, 6, 2, 10, 0, 0, 0, time.UTC),
			want:     false,
		},
		{
			name:     "post scheduled for yesterday should return false",
			postTime: time.Date(2024, 5, 31, 15, 0, 0, 0, time.UTC),
			want:     false,
		},
		{
			name:     "post scheduled for exact current time should return false",
			postTime: currentTime,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTodayAndFuture(tt.postTime, currentTime)
			if got != tt.want {
				t.Errorf("IsTodayAndFuture() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsToday(t *testing.T) {
	currentTime := time.Date(2024, 6, 1, 12, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		postTime time.Time
		want     bool
	}{
		{
			name:     "post scheduled for later today should return true",
			postTime: time.Date(2024, 6, 1, 15, 0, 0, 0, time.UTC),
			want:     true,
		},
		{
			name:     "post scheduled for earlier today should return true",
			postTime: time.Date(2024, 6, 1, 8, 0, 0, 0, time.UTC),
			want:     true,
		},
		{
			name:     "post scheduled for tomorrow should return false",
			postTime: time.Date(2024, 6, 2, 10, 0, 0, 0, time.UTC),
			want:     false,
		},
		{
			name:     "post scheduled for yesterday should return false",
			postTime: time.Date(2024, 5, 31, 15, 0, 0, 0, time.UTC),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsToday(tt.postTime, currentTime)
			if got != tt.want {
				t.Errorf("IsToday() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterFuturePosts(t *testing.T) {
	currentTime := time.Date(2024, 6, 1, 12, 30, 0, 0, time.UTC)
	pastTime := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)
	futureTime := time.Date(2024, 6, 1, 15, 0, 0, 0, time.UTC)
	tomorrowTime := time.Date(2024, 6, 2, 10, 0, 0, 0, time.UTC)

	posts := []config.Post{
		{
			Content:     "Past post",
			ScheduledAt: pastTime,
			Enabled:     true,
		},
		{
			Content:     "Future post",
			ScheduledAt: futureTime,
			Enabled:     true,
		},
		{
			Content:     "Tomorrow post",
			ScheduledAt: tomorrowTime,
			Enabled:     true,
		},
		{
			Content:     "Test post",
			ScheduledAt: tomorrowTime,
			Enabled:     true,
			Test:        true,
		},
		{
			Content:     "Disabled post",
			ScheduledAt: futureTime,
			Enabled:     false,
		},
	}

	filtered := FilterFuturePosts(posts, currentTime)

	// Should have 2 posts: future post and test post
	if len(filtered) != 2 {
		t.Errorf("FilterFuturePosts() returned %d posts, want 2", len(filtered))
	}

	// Check that we got the right posts
	expectedContents := map[string]bool{
		"Future post": false,
		"Test post":   false,
	}

	for _, post := range filtered {
		if _, exists := expectedContents[post.Content]; exists {
			expectedContents[post.Content] = true
		} else {
			t.Errorf("FilterFuturePosts() returned unexpected post: %s", post.Content)
		}
	}

	for content, found := range expectedContents {
		if !found {
			t.Errorf("FilterFuturePosts() missing expected post: %s", content)
		}
	}
}

func TestSortByExecuteTime(t *testing.T) {
	posts := []ScheduledPost{
		{
			Post:      config.Post{Content: "Third post"},
			ExecuteAt: time.Date(2024, 6, 1, 15, 0, 0, 0, time.UTC),
		},
		{
			Post:      config.Post{Content: "First post"},
			ExecuteAt: time.Date(2024, 6, 1, 8, 0, 0, 0, time.UTC),
		},
		{
			Post:      config.Post{Content: "Second post"},
			ExecuteAt: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		},
	}

	sorted := SortByExecuteTime(posts)

	expectedOrder := []string{"First post", "Second post", "Third post"}

	if len(sorted) != len(expectedOrder) {
		t.Errorf("SortByExecuteTime() returned %d posts, want %d", len(sorted), len(expectedOrder))
	}

	for i, expectedContent := range expectedOrder {
		if sorted[i].Post.Content != expectedContent {
			t.Errorf("SortByExecuteTime()[%d].Post.Content = %v, want %v", i, sorted[i].Post.Content, expectedContent)
		}
	}
}
