package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Load reads and parses the YAML configuration file
func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

// Checks the configuration for errors
func (c *Config) Validate() error {
	if len(c.Posts) == 0 {
		return fmt.Errorf("no posts configured")
	}

	now := time.Now()
	pastPostCount := 0

	for i, post := range c.Posts {
		if post.Content == "" {
			return fmt.Errorf("post %d: content is required", i)
		}
		if post.ScheduledAt.IsZero() {
			return fmt.Errorf("post %d: scheduled_at is required", i)
		}

		// Count past posts but don't fail validation
		if post.ScheduledAt.Before(now) {
			pastPostCount++
			if post.Enabled {
				// Only warn about enabled posts in the past
				fmt.Printf("Warning: post %d is scheduled in the past but enabled: %s\n",
					i, post.ScheduledAt.Format("2006-01-02 15:04"))
			}
		}
	}

	if pastPostCount > 0 {
		fmt.Printf("Info: %d posts are scheduled in the past (consider disabling them)\n", pastPostCount)
	}

	return nil
}

// Returns the API token from environment variable or config
func (c *Config) GetAPIToken() string {
	if token := os.Getenv("X_BEARER_TOKEN"); token != "" {
		return token
	}
	if c.API != nil {
		return c.API.BearerToken
	}
	return ""
}

// Returns only enabled posts
func (c *Config) GetEnabledPosts() []Post {
	var enabled []Post
	for _, post := range c.Posts {
		if post.Enabled {
			enabled = append(enabled, post)
		}
	}
	return enabled
}

// Returns posts scheduled in the future (excluding test posts)
func (c *Config) GetFuturePosts() []Post {
	now := time.Now()
	var future []Post
	for _, post := range c.GetEnabledPosts() {
		// Test posts are not included in cron schedule
		if !post.Test && post.ScheduledAt.After(now) {
			future = append(future, post)
		}
	}
	return future
}
