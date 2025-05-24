package xapi

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/zinrai/x-scheduler/pkg/logger"
)

// Defines retry behavior
type RetryConfig struct {
	MaxRetries    int
	BaseDelay     time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

// Returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		BaseDelay:     1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
	}
}

// Determines if an error should trigger a retry
func ShouldRetry(statusCode int, err error) bool {
	// Network errors should be retried
	if err != nil {
		return true
	}

	// Retry on server errors
	if statusCode >= 500 {
		return true
	}

	// Retry on rate limiting
	if statusCode == 429 {
		return true
	}

	// Don't retry client errors
	if statusCode >= 400 && statusCode < 500 {
		return false
	}

	return false
}

// Calculates the delay for the next retry attempt
func (rc *RetryConfig) CalculateDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return rc.BaseDelay
	}

	// Exponential backoff: BaseDelay * (BackoffFactor ^ attempt)
	delay := time.Duration(float64(rc.BaseDelay) * math.Pow(rc.BackoffFactor, float64(attempt)))

	// Cap at MaxDelay
	if delay > rc.MaxDelay {
		delay = rc.MaxDelay
	}

	return delay
}

// Executes a function with retry logic
func (rc *RetryConfig) ExecuteWithRetry(operation func() (*http.Response, error)) (*http.Response, error) {
	var lastErr error
	var resp *http.Response

	for attempt := 0; attempt <= rc.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := rc.CalculateDelay(attempt - 1)
			logger.Info("Retrying in %v (attempt %d/%d)", delay, attempt, rc.MaxRetries)
			time.Sleep(delay)
		}

		resp, lastErr = operation()

		// Success case - early return
		if rc.isSuccessfulResponse(resp, lastErr) {
			if attempt > 0 {
				logger.Info("Request succeeded after %d retries", attempt)
			}
			return resp, nil
		}

		// Check if we should retry
		statusCode := rc.getStatusCode(resp)
		if !ShouldRetry(statusCode, lastErr) {
			logger.Debug("Not retrying - error type not retryable (status: %d)", statusCode)
			break
		}

		// Log the retry attempt
		rc.logRetryAttempt(lastErr, statusCode, attempt)

		// Clean up response body to prevent resource leak
		rc.closeResponseBody(resp)
	}

	// All retries exhausted - return appropriate error
	return rc.handleRetryFailure(resp, lastErr)
}

// Checks if the response indicates success
func (rc *RetryConfig) isSuccessfulResponse(resp *http.Response, err error) bool {
	return err == nil && resp != nil && resp.StatusCode < 400
}

// Safely extracts status code from response
func (rc *RetryConfig) getStatusCode(resp *http.Response) int {
	if resp != nil {
		return resp.StatusCode
	}
	return 0
}

// Logs retry information
func (rc *RetryConfig) logRetryAttempt(err error, statusCode, attempt int) {
	if err != nil {
		logger.Warn("Request failed with error: %v (attempt %d/%d)", err, attempt+1, rc.MaxRetries+1)
		return
	}

	logger.Warn("Request failed with status %d (attempt %d/%d)", statusCode, attempt+1, rc.MaxRetries+1)
}

// Safely closes response body
func (rc *RetryConfig) closeResponseBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
}

// Returns appropriate error after all retries exhausted
func (rc *RetryConfig) handleRetryFailure(resp *http.Response, lastErr error) (*http.Response, error) {
	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", rc.MaxRetries, lastErr)
	}

	return resp, fmt.Errorf("request failed after %d retries with status %d", rc.MaxRetries, resp.StatusCode)
}
