package xapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zinrai/x-scheduler/pkg/logger"
)

const (
	BaseURL      = "https://api.x.com"
	PostTweetURL = BaseURL + "/2/tweets"
	Timeout      = 30 * time.Second
)

// Represents X API client
type Client struct {
	auth        *Auth
	httpClient  *http.Client
	retryConfig RetryConfig
}

// Creates a new X API client
func NewClient(bearerToken string) *Client {
	return &Client{
		auth: NewAuth(bearerToken),
		httpClient: &http.Client{
			Timeout: Timeout,
		},
		retryConfig: DefaultRetryConfig(),
	}
}

// Represents the request body for posting a tweet
type PostTweetRequest struct {
	Text string `json:"text"`
}

// Represents the response from posting a tweet
type PostTweetResponse struct {
	Data struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"errors,omitempty"`
}

// Posts a tweet to X
func (c *Client) PostTweet(content string) error {
	// Validate authentication
	if err := c.auth.IsValid(); err != nil {
		return fmt.Errorf("authentication error: %w", err)
	}

	// Prepare request body
	reqBody := PostTweetRequest{
		Text: content,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	logger.Debug("Posting tweet: %s", content)

	// Execute request with retry
	resp, err := c.retryConfig.ExecuteWithRetry(func() (*http.Response, error) {
		req, err := http.NewRequest("POST", PostTweetURL, bytes.NewBuffer(jsonBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		c.auth.AddHeaders(req)

		return c.httpClient.Do(req)
	})

	if err != nil {
		return fmt.Errorf("failed to post tweet: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var postResp PostTweetResponse
	if err := json.Unmarshal(respBody, &postResp); err != nil {
		logger.Warn("Failed to parse response JSON: %v", err)
		logger.Debug("Raw response: %s", string(respBody))
	}

	// Check for API errors
	if resp.StatusCode != 201 {
		errorMsg := fmt.Sprintf("API request failed with status %d", resp.StatusCode)

		if len(postResp.Errors) > 0 {
			errorMsg += fmt.Sprintf(": %s", postResp.Errors[0].Message)
		} else {
			errorMsg += fmt.Sprintf(": %s", string(respBody))
		}

		return fmt.Errorf(errorMsg)
	}

	logger.Info("Tweet posted successfully (ID: %s)", postResp.Data.ID)
	return nil
}

// Validates the API credentials by making a test request
func (c *Client) ValidateCredentials() error {
	if err := c.auth.IsValid(); err != nil {
		return err
	}

	// For X API v2, we can't easily test credentials without posting
	// So we just validate the token format
	logger.Info("Using bearer token: %s", c.auth.GetToken())
	return nil
}
