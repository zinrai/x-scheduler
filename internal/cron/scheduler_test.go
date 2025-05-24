package cron

import (
	"testing"
	"time"
)

func TestTimeToCron(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "basic time conversion",
			// Use Local timezone instead of UTC
			time: time.Date(2024, 6, 15, 14, 30, 0, 0, time.Local),
			want: "30 14 15 6 *",
		},
		{
			name: "beginning of year",
			time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
			want: "0 0 1 1 *",
		},
		{
			name: "end of year",
			time: time.Date(2024, 12, 31, 23, 59, 0, 0, time.Local),
			want: "59 23 31 12 *",
		},
		{
			name: "leap year february",
			time: time.Date(2024, 2, 29, 12, 15, 0, 0, time.Local),
			want: "15 12 29 2 *",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TimeToCron(tt.time)
			if got != tt.want {
				t.Errorf("TimeToCron() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeToCronWithTimezone(t *testing.T) {
	// Test timezone handling
	jst, _ := time.LoadLocation("Asia/Tokyo")
	utcTime := time.Date(2024, 6, 15, 5, 30, 0, 0, time.UTC)
	jstTime := utcTime.In(jst) // Should be 14:30 JST

	cronStr := TimeToCron(jstTime)
	expected := "30 14 15 6 *"

	if cronStr != expected {
		t.Errorf("TimeToCron() with timezone = %v, want %v", cronStr, expected)
	}
}

func TestFormatCronEntry(t *testing.T) {
	tests := []struct {
		name  string
		entry CronEntry
		want  string
	}{
		{
			name: "entry with comment",
			entry: CronEntry{
				Schedule: "30 14 15 6 *",
				Command:  "/usr/bin/x-scheduler -execute config.yaml",
				Comment:  "2024-06-15 14:30 - Good morning post",
			},
			want: "# 2024-06-15 14:30 - Good morning post\n30 14 15 6 * root /usr/bin/x-scheduler -execute config.yaml",
		},
		{
			name: "entry without comment",
			entry: CronEntry{
				Schedule: "0 9 1 * *",
				Command:  "/usr/bin/x-scheduler -execute config.yaml",
				Comment:  "",
			},
			want: "0 9 1 * * root /usr/bin/x-scheduler -execute config.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatCronEntry(tt.entry)
			if got != tt.want {
				t.Errorf("FormatCronEntry() = %v, want %v", got, tt.want)
			}
		})
	}
}
