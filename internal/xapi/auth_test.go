package xapi

import (
	"net/http"
	"testing"
)

func TestAuth_IsValid(t *testing.T) {
	tests := []struct {
		name        string
		bearerToken string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "empty token should return error",
			bearerToken: "",
			wantErr:     true,
			errMsg:      "bearer token is required",
		},
		{
			name:        "short token should return error",
			bearerToken: "short",
			wantErr:     true,
			errMsg:      "bearer token appears to be invalid",
		},
		{
			name:        "valid token should pass",
			bearerToken: "valid_bearer_token_12345",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewAuth(tt.bearerToken)
			err := auth.IsValid()

			if tt.wantErr {
				if err == nil {
					t.Errorf("IsValid() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("IsValid() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("IsValid() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestAuth_AddHeaders(t *testing.T) {
	token := "test_bearer_token_12345"
	auth := NewAuth(token)

	req, _ := http.NewRequest("POST", "https://api.x.com/2/tweets", nil)
	auth.AddHeaders(req)

	// Check Authorization header
	expectedAuth := "Bearer " + token
	if got := req.Header.Get("Authorization"); got != expectedAuth {
		t.Errorf("Authorization header = %v, want %v", got, expectedAuth)
	}

	// Check Content-Type header
	expectedContentType := "application/json"
	if got := req.Header.Get("Content-Type"); got != expectedContentType {
		t.Errorf("Content-Type header = %v, want %v", got, expectedContentType)
	}

	// Check User-Agent header
	expectedUserAgent := "x-scheduler/1.0"
	if got := req.Header.Get("User-Agent"); got != expectedUserAgent {
		t.Errorf("User-Agent header = %v, want %v", got, expectedUserAgent)
	}
}
