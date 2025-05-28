package executor

import (
	"fmt"
	"sort"
	"time"

	"github.com/zinrai/x-scheduler/internal/config"
	"github.com/zinrai/x-scheduler/internal/poster"
	"github.com/zinrai/x-scheduler/pkg/logger"
)

// Represents a post with its execution time
type ScheduledPost struct {
	Post      config.Post
	ExecuteAt time.Time
}

// Handles the execution of scheduled posts
type Executor struct {
	jobQueue chan ScheduledPost
}

// Creates a new executor instance
func NewExecutor() *Executor {
	return &Executor{
		jobQueue: make(chan ScheduledPost, 100), // Buffer for up to 100 posts
	}
}

// Processes all posts scheduled for today that are in the future
func (e *Executor) Execute(cfg *config.Config) error {
	logger.Info("Starting execution")

	// Validate poster (xurl command)
	if err := poster.Validate(); err != nil {
		return fmt.Errorf("poster validation failed: %w", err)
	}

	// Get future posts for today
	futurePosts := e.getFuturePosts(cfg)
	if len(futurePosts) == 0 {
		logger.Info("No posts scheduled for execution")
		return nil
	}

	logger.Info("Found %d posts scheduled for execution", len(futurePosts))

	// Sort posts by execution time
	sort.Slice(futurePosts, func(i, j int) bool {
		return futurePosts[i].ExecuteAt.Before(futurePosts[j].ExecuteAt)
	})

	// Queue all posts
	e.queuePosts(futurePosts)

	// Process queue sequentially
	return e.processQueue()
}

// Returns posts scheduled for today that are in the future
func (e *Executor) getFuturePosts(cfg *config.Config) []ScheduledPost {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)

	var futurePosts []ScheduledPost

	for _, post := range cfg.GetEnabledPosts() {
		// Test posts are executed immediately regardless of schedule
		if post.Test {
			futurePosts = append(futurePosts, ScheduledPost{
				Post:      post,
				ExecuteAt: now, // Execute immediately
			})
			continue
		}

		// Check if post is scheduled for today
		if post.ScheduledAt.After(today) && post.ScheduledAt.Before(tomorrow) {
			// Only include future posts
			if post.ScheduledAt.After(now) {
				futurePosts = append(futurePosts, ScheduledPost{
					Post:      post,
					ExecuteAt: post.ScheduledAt,
				})
			} else {
				// Log skipped past posts
				logger.Info("Skipping past post: %s (scheduled at %s)",
					truncateContent(post.Content, 30),
					post.ScheduledAt.Format("15:04:05"))
			}
		}
	}

	return futurePosts
}

// Adds posts to the processing queue
func (e *Executor) queuePosts(posts []ScheduledPost) {
	for _, scheduledPost := range posts {
		nextPostTime := time.Until(scheduledPost.ExecuteAt)
		if nextPostTime > 0 {
			logger.Info("Queuing post: %s (in %v at %s)",
				truncateContent(scheduledPost.Post.Content, 30),
				nextPostTime.Round(time.Second),
				scheduledPost.ExecuteAt.Format("15:04:05"))
		} else {
			logger.Info("Queuing immediate post: %s",
				truncateContent(scheduledPost.Post.Content, 30))
		}

		e.jobQueue <- scheduledPost
	}

	// Close the queue to signal no more posts
	close(e.jobQueue)
}

// Processes posts from the queue sequentially
func (e *Executor) processQueue() error {
	var errors []error
	successCount := 0

	for scheduledPost := range e.jobQueue {
		// Wait until it's time to post
		e.waitUntilTime(scheduledPost.ExecuteAt)

		// Execute the post
		if err := e.executePost(scheduledPost); err != nil {
			logger.Error("Failed to execute post: %v", err)
			errors = append(errors, err)
		} else {
			successCount++
		}
	}

	// Report results
	logger.Info("Execution completed: %d successful, %d failed", successCount, len(errors))

	if len(errors) > 0 {
		return fmt.Errorf("some posts failed: %v", errors)
	}

	return nil
}

// Waits until the specified time
func (e *Executor) waitUntilTime(executeAt time.Time) {
	now := time.Now()
	if executeAt.After(now) {
		waitDuration := executeAt.Sub(now)
		logger.Debug("Waiting %v until execution time (%s)", waitDuration, executeAt.Format("15:04:05"))
		time.Sleep(waitDuration)
	}
}

// Executes a single post
func (e *Executor) executePost(scheduledPost ScheduledPost) error {
	post := scheduledPost.Post

	// Handle dry run
	if post.DryRun {
		logger.Info("DRY RUN: Would post: %s", post.Content)
		fmt.Printf("✓ [DRY RUN] Would post: %s\n", post.Content)
		return nil
	}

	// Handle test posts
	if post.Test {
		logger.Info("Test post: %s", truncateContent(post.Content, 50))
	} else {
		logger.Info("Posting: %s", truncateContent(post.Content, 50))
	}

	// Execute actual post
	if err := poster.Post(post.Content); err != nil {
		return fmt.Errorf("failed to post '%s': %w",
			truncateContent(post.Content, 30), err)
	}

	// Success message
	if post.Test {
		fmt.Printf("✓ Test post successful: %s\n", truncateContent(post.Content, 50))
	} else {
		logger.Info("Post successful: %s", truncateContent(post.Content, 50))
	}

	return nil
}

// Returns information about scheduled posts
func (e *Executor) GetStatus(cfg *config.Config) (map[string]interface{}, error) {
	enabledPosts := cfg.GetEnabledPosts()
	futurePosts := e.getFuturePosts(cfg)

	status := map[string]interface{}{
		"total_posts":   len(cfg.Posts),
		"enabled_posts": len(enabledPosts),
		"future_posts":  len(futurePosts),
		"current_time":  time.Now().Format(time.RFC3339),
	}

	if len(futurePosts) > 0 {
		// Find next post
		sort.Slice(futurePosts, func(i, j int) bool {
			return futurePosts[i].ExecuteAt.Before(futurePosts[j].ExecuteAt)
		})
		nextPost := futurePosts[0]
		status["next_post_time"] = nextPost.ExecuteAt.Format(time.RFC3339)
		status["next_post_in"] = time.Until(nextPost.ExecuteAt).String()
	}

	return status, nil
}

// Truncates content for logging
func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen-3] + "..."
}
