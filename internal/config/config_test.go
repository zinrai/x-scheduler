package config

import (
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "empty posts should return error",
			config: Config{
				Posts: []Post{},
			},
			wantErr: true,
			errMsg:  "no posts configured",
		},
		{
			name: "post with empty content should return error",
			config: Config{
				Posts: []Post{
					{
						Content:     "",
						ScheduledAt: time.Now().Add(time.Hour),
						Enabled:     true,
					},
				},
			},
			wantErr: true,
			errMsg:  "post 0: content is required",
		},
		{
			name: "post with zero scheduled time should return error",
			config: Config{
				Posts: []Post{
					{
						Content:     "Test content",
						ScheduledAt: time.Time{},
						Enabled:     true,
					},
				},
			},
			wantErr: true,
			errMsg:  "post 0: scheduled_at is required",
		},
		{
			name: "valid config should pass validation",
			config: Config{
				Posts: []Post{
					{
						Content:     "Test content",
						ScheduledAt: time.Now().Add(time.Hour),
						Enabled:     true,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestConfig_GetEnabledPosts(t *testing.T) {
	config := Config{
		Posts: []Post{
			{Content: "Post 1", Enabled: true},
			{Content: "Post 2", Enabled: false},
			{Content: "Post 3", Enabled: true},
			{Content: "Post 4"}, // enabled defaults to false
		},
	}

	enabled := config.GetEnabledPosts()

	if len(enabled) != 2 {
		t.Errorf("GetEnabledPosts() returned %d posts, want 2", len(enabled))
	}

	expectedContents := []string{"Post 1", "Post 3"}
	for i, post := range enabled {
		if post.Content != expectedContents[i] {
			t.Errorf("GetEnabledPosts()[%d].Content = %v, want %v", i, post.Content, expectedContents[i])
		}
	}
}

func TestConfig_GetFuturePosts(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-time.Hour)
	futureTime := now.Add(time.Hour)

	config := Config{
		Posts: []Post{
			{Content: "Past post", ScheduledAt: pastTime, Enabled: true},
			{Content: "Future post", ScheduledAt: futureTime, Enabled: true},
			{Content: "Test post", ScheduledAt: futureTime, Enabled: true, Test: true},
			{Content: "Disabled future post", ScheduledAt: futureTime, Enabled: false},
		},
	}

	future := config.GetFuturePosts()

	if len(future) != 1 {
		t.Errorf("GetFuturePosts() returned %d posts, want 1", len(future))
	}

	if future[0].Content != "Future post" {
		t.Errorf("GetFuturePosts()[0].Content = %v, want 'Future post'", future[0].Content)
	}
}
