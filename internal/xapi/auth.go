package xapi

import (
	"fmt"
	"net/http"
)

// Handles X API authentication
type Auth struct {
	bearerToken string
}

// Creates a new Auth instance
func NewAuth(bearerToken string) *Auth {
	return &Auth{
		bearerToken: bearerToken,
	}
}

// Checks if the authentication is properly configured
func (a *Auth) IsValid() error {
	if a.bearerToken == "" {
		return fmt.Errorf("bearer token is required")
	}

	// Basic format validation
	if len(a.bearerToken) < 10 {
		return fmt.Errorf("bearer token appears to be invalid")
	}

	return nil
}

// Adds authentication headers to HTTP request
func (a *Auth) AddHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+a.bearerToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "x-scheduler/1.0")
}

// Returns the bearer token (for debugging purposes)
func (a *Auth) GetToken() string {
	if len(a.bearerToken) <= 8 {
		return a.bearerToken
	}
	// Return masked token for security
	return a.bearerToken[:4] + "..." + a.bearerToken[len(a.bearerToken)-4:]
}
