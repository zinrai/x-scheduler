package cron

import (
	"fmt"
	"time"
)

// Represents a single cron job entry
type CronEntry struct {
	Schedule string
	Command  string
	Comment  string
}

// Converts time.Time to cron format (minute hour day month weekday)
func TimeToCron(t time.Time) string {
	// Convert to local time
	local := t.Local()

	return fmt.Sprintf("%d %d %d %d *",
		local.Minute(),
		local.Hour(),
		local.Day(),
		int(local.Month()),
	)
}

// Formats a cron entry for output
func FormatCronEntry(entry CronEntry) string {
	result := ""
	if entry.Comment != "" {
		result += fmt.Sprintf("# %s\n", entry.Comment)
	}
	result += fmt.Sprintf("%s root %s", entry.Schedule, entry.Command)
	return result
}

// Truncates long comments to reasonable length
func TruncateComment(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen-3] + "..."
}
