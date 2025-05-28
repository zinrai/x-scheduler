package executor

import (
	"time"

	"github.com/zinrai/x-scheduler/internal/config"
)

// Checks if the given time is today and in the future
func IsTodayAndFuture(postTime, currentTime time.Time) bool {
	today := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
	tomorrow := today.Add(24 * time.Hour)

	// Check if post is today and in the future
	return postTime.After(today) && postTime.Before(tomorrow) && postTime.After(currentTime)
}

// Checks if the given time is today
func IsToday(postTime, currentTime time.Time) bool {
	today := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
	tomorrow := today.Add(24 * time.Hour)

	return postTime.After(today) && postTime.Before(tomorrow)
}

// Filters posts to include only future posts for today
func FilterFuturePosts(posts []config.Post, currentTime time.Time) []config.Post {
	var futurePosts []config.Post

	for _, post := range posts {
		if post.Enabled {
			// Test posts are always included
			if post.Test {
				futurePosts = append(futurePosts, post)
				continue
			}

			// Regular posts must be today and in the future
			if IsTodayAndFuture(post.ScheduledAt, currentTime) {
				futurePosts = append(futurePosts, post)
			}
		}
	}

	return futurePosts
}

// Sorts scheduled posts by execution time
func SortByExecuteTime(posts []ScheduledPost) []ScheduledPost {
	sorted := make([]ScheduledPost, len(posts))
	copy(sorted, posts)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].ExecuteAt.Before(sorted[i].ExecuteAt) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// Returns the next scheduled post time from the given posts
func GetNextScheduledTime(posts []config.Post) *time.Time {
	var next *time.Time
	now := time.Now()

	for _, post := range posts {
		if post.Enabled && post.ScheduledAt.After(now) {
			if next == nil || post.ScheduledAt.Before(*next) {
				next = &post.ScheduledAt
			}
		}
	}

	return next
}

// Returns the number of posts scheduled in the future for today
func CountFuturePosts(posts []config.Post) int {
	now := time.Now()
	count := 0

	for _, post := range posts {
		if post.Enabled {
			if post.Test || IsTodayAndFuture(post.ScheduledAt, now) {
				count++
			}
		}
	}

	return count
}
