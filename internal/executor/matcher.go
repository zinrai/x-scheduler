package executor

import (
	"time"

	"github.com/zinrai/x-scheduler/internal/config"
)

// Defines the acceptable time window for matching scheduled posts
const TimeWindow = 1 * time.Minute

// Finds posts that should be posted at the current time
func FindMatchingPosts(posts []config.Post, currentTime time.Time) []config.Post {
	var matches []config.Post

	for _, post := range posts {
		// Test posts are always executed regardless of schedule
		if post.Test {
			matches = append(matches, post)
		} else if ShouldPostNow(post.ScheduledAt, currentTime) {
			matches = append(matches, post)
		}
	}

	return matches
}

// Returns posts marked for testing
func FindTestPosts(posts []config.Post) []config.Post {
	var testPosts []config.Post

	for _, post := range posts {
		if post.Test {
			testPosts = append(testPosts, post)
		}
	}

	return testPosts
}

// Determines if a post should be posted at the current time
func ShouldPostNow(scheduledAt, currentTime time.Time) bool {
	// Truncate both times to minute precision for comparison
	scheduledMinute := scheduledAt.Truncate(time.Minute)
	currentMinute := currentTime.Truncate(time.Minute)

	// Allow posting within the time window
	diff := currentMinute.Sub(scheduledMinute)
	if diff < 0 {
		diff = -diff
	}

	return diff <= TimeWindow
}

// Returns the next scheduled post time
func GetNextScheduledTime(posts []config.Post) *time.Time {
	var next *time.Time
	now := time.Now()

	for _, post := range posts {
		if post.ScheduledAt.After(now) {
			if next == nil || post.ScheduledAt.Before(*next) {
				next = &post.ScheduledAt
			}
		}
	}

	return next
}

// Returns the number of posts scheduled in the future
func CountPendingPosts(posts []config.Post) int {
	count := 0
	now := time.Now()

	for _, post := range posts {
		if post.ScheduledAt.After(now) {
			count++
		}
	}

	return count
}
