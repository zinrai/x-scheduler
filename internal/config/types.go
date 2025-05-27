package config

import "time"

// Represents the complete configuration structure
type Config struct {
	Posts []Post `yaml:"posts"`
}

// Represents a single scheduled post
type Post struct {
	Content     string    `yaml:"content"`
	ScheduledAt time.Time `yaml:"scheduled_at"`
	Enabled     bool      `yaml:"enabled"`
	Test        bool      `yaml:"test,omitempty"`    // Execute immediately for testing
	DryRun      bool      `yaml:"dry_run,omitempty"` // Don't actually post (test mode only)
}
