package xapi

import (
	"errors"
	"testing"
	"time"
)

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		err        error
		want       bool
	}{
		{
			name:       "5xx server error should retry",
			statusCode: 500,
			err:        nil,
			want:       true,
		},
		{
			name:       "502 bad gateway should retry",
			statusCode: 502,
			err:        nil,
			want:       true,
		},
		{
			name:       "429 rate limit should retry",
			statusCode: 429,
			err:        nil,
			want:       true,
		},
		{
			name:       "4xx client error should not retry",
			statusCode: 400,
			err:        nil,
			want:       false,
		},
		{
			name:       "401 unauthorized should not retry",
			statusCode: 401,
			err:        nil,
			want:       false,
		},
		{
			name:       "404 not found should not retry",
			statusCode: 404,
			err:        nil,
			want:       false,
		},
		{
			name:       "network error should retry",
			statusCode: 0,
			err:        errors.New("network error"),
			want:       true,
		},
		{
			name:       "200 success should not retry",
			statusCode: 200,
			err:        nil,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldRetry(tt.statusCode, tt.err)
			if got != tt.want {
				t.Errorf("ShouldRetry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRetryConfig_CalculateDelay(t *testing.T) {
	config := DefaultRetryConfig()

	tests := []struct {
		name     string
		attempt  int
		want     time.Duration
		maxCheck bool // Check if result is capped at MaxDelay
	}{
		{
			name:    "first retry should be base delay * backoff factor",
			attempt: 1,
			want:    2 * time.Second, // BaseDelay(1s) * BackoffFactor(2) ^ 1
		},
		{
			name:    "second retry should be exponentially increased",
			attempt: 2,
			want:    4 * time.Second, // BaseDelay(1s) * BackoffFactor(2) ^ 2
		},
		{
			name:    "third retry should be further exponentially increased",
			attempt: 3,
			want:    8 * time.Second, // BaseDelay(1s) * BackoffFactor(2) ^ 3
		},
		{
			name:     "large attempt should be capped at max delay",
			attempt:  10,
			maxCheck: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.CalculateDelay(tt.attempt)

			if tt.maxCheck {
				if got > config.MaxDelay {
					t.Errorf("CalculateDelay() = %v, should not exceed MaxDelay %v", got, config.MaxDelay)
				}
			} else {
				if got != tt.want {
					t.Errorf("CalculateDelay() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestRetryConfig_CalculateDelayEdgeCases(t *testing.T) {
	config := DefaultRetryConfig()

	// Test zero or negative attempt
	delay := config.CalculateDelay(0)
	if delay != config.BaseDelay {
		t.Errorf("CalculateDelay(0) = %v, want %v", delay, config.BaseDelay)
	}

	delay = config.CalculateDelay(-1)
	if delay != config.BaseDelay {
		t.Errorf("CalculateDelay(-1) = %v, want %v", delay, config.BaseDelay)
	}
}
