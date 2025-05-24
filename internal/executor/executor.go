package executor

import (
	"fmt"
	"time"

	"github.com/zinrai/x-scheduler/internal/config"
	"github.com/zinrai/x-scheduler/internal/xapi"
	"github.com/zinrai/x-scheduler/pkg/logger"
)

// Handles the execution of scheduled posts
type Executor struct {
	client *xapi.Client
}

// Creates a new executor instance
func NewExecutor(bearerToken string) *Executor {
	return &Executor{
		client: xapi.NewClient(bearerToken),
	}
}

// Finds and posts tweets that should be posted at the current time
func (e *Executor) Execute(cfg *config.Config) error {
	logger.Info("Starting execution check")

	// Validate API credentials
	if err := e.client.ValidateCredentials(); err != nil {
		return fmt.Errorf("invalid API credentials: %w", err)
	}

	// Get enabled posts
	enabledPosts := cfg.GetEnabledPosts()
	logger.Debug("Found %d enabled posts", len(enabledPosts))

	// Find posts that should be posted now (including test posts)
	currentTime := time.Now()
	matchingPosts := FindMatchingPosts(enabledPosts, currentTime)

	if len(matchingPosts) == 0 {
		logger.Info("No posts scheduled for current time (%s)", currentTime.Format("2006-01-02 15:04"))
		return nil
	}

	logger.Info("Found %d posts to publish", len(matchingPosts))

	// Separate test posts and regular posts for different handling
	testPosts, regularPosts := e.separatePostsByType(matchingPosts)

	// Post each matching tweet
	var errors []error
	successCount := 0

	// Handle regular posts
	successCount += e.processRegularPosts(regularPosts, &errors)

	// Handle test posts
	successCount += e.processTestPosts(testPosts, &errors)

	// Report results
	logger.Info("Execution completed: %d successful, %d failed", successCount, len(errors))

	if len(errors) > 0 {
		return fmt.Errorf("some posts failed: %v", errors)
	}

	return nil
}

// Separates posts into test and regular posts
func (e *Executor) separatePostsByType(posts []config.Post) ([]config.Post, []config.Post) {
	var testPosts, regularPosts []config.Post
	for _, post := range posts {
		if post.Test {
			testPosts = append(testPosts, post)
		} else {
			regularPosts = append(regularPosts, post)
		}
	}
	return testPosts, regularPosts
}

// Handles regular scheduled posts
func (e *Executor) processRegularPosts(posts []config.Post, errors *[]error) int {
	successCount := 0

	for i, post := range posts {
		logger.Info("Posting tweet %d/%d: %s", i+1, len(posts), truncateContent(post.Content, 50))

		if err := e.client.PostTweet(post.Content); err != nil {
			logger.Error("Failed to post tweet: %v", err)
			*errors = append(*errors, fmt.Errorf("failed to post tweet '%s': %w", truncateContent(post.Content, 30), err))
			continue
		}

		successCount++
	}

	return successCount
}

// Handles test posts (with dry-run support)
func (e *Executor) processTestPosts(posts []config.Post, errors *[]error) int {
	successCount := 0

	for i, post := range posts {
		if post.DryRun {
			successCount += e.processDryRunPost(post, i, len(posts))
			continue
		}

		successCount += e.processActualTestPost(post, i, len(posts), errors)
	}

	return successCount
}

// Handles a single dry-run test post
func (e *Executor) processDryRunPost(post config.Post, index, total int) int {
	logger.Info("Test post %d/%d (DRY RUN): %s", index+1, total, truncateContent(post.Content, 50))
	fmt.Printf("[DRY RUN] Would post: %s\n", post.Content)
	return 1
}

// Handles a single actual test post
func (e *Executor) processActualTestPost(post config.Post, index, total int, errors *[]error) int {
	logger.Info("Test post %d/%d: %s", index+1, total, truncateContent(post.Content, 50))

	if err := e.client.PostTweet(post.Content); err != nil {
		logger.Error("Failed to post test tweet: %v", err)
		*errors = append(*errors, fmt.Errorf("failed to post test tweet '%s': %w", truncateContent(post.Content, 30), err))
		return 0
	}

	fmt.Printf("✓ Test post successful: %s\n", truncateContent(post.Content, 50))
	return 1
}

// Returns information about scheduled posts
func (e *Executor) GetStatus(cfg *config.Config) (map[string]interface{}, error) {
	enabledPosts := cfg.GetEnabledPosts()
	futurePosts := cfg.GetFuturePosts()

	status := map[string]interface{}{
		"total_posts":   len(cfg.Posts),
		"enabled_posts": len(enabledPosts),
		"future_posts":  len(futurePosts),
		"current_time":  time.Now().Format(time.RFC3339),
	}

	if next := GetNextScheduledTime(enabledPosts); next != nil {
		status["next_post_time"] = next.Format(time.RFC3339)
		status["next_post_in"] = time.Until(*next).String()
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
