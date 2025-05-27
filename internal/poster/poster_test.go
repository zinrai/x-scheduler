package poster

import (
	"encoding/json"
	"testing"
)

// Test JSON marshaling behavior used in Post function
func TestJSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "Hello world!",
			expected: `{"text":"Hello world!"}`,
		},
		{
			name:     "text with quotes",
			input:    `Say "hello" to the world`,
			expected: `{"text":"Say \"hello\" to the world"}`,
		},
		{
			name:     "text with newlines",
			input:    "Line 1\nLine 2",
			expected: `{"text":"Line 1\nLine 2"}`,
		},
		{
			name:     "text with tabs",
			input:    "Column1\tColumn2",
			expected: `{"text":"Column1\tColumn2"}`,
		},
		{
			name:     "text with backslashes",
			input:    `Path: C:\Users\Name`,
			expected: `{"text":"Path: C:\\Users\\Name"}`,
		},
		{
			name:     "text with unicode",
			input:    "Hello 世界! 🌍",
			expected: `{"text":"Hello 世界! 🌍"}`,
		},
		{
			name:     "empty text",
			input:    "",
			expected: `{"text":""}`,
		},
		{
			name:     "text with control characters",
			input:    "Line1\r\nLine2\tTab\b\f",
			expected: `{"text":"Line1\r\nLine2\tTab\b\f"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the same JSON marshaling logic used in Post function
			reqBody := struct {
				Text string `json:"text"`
			}{
				Text: tt.input,
			}

			jsonBytes, err := json.Marshal(reqBody)
			if err != nil {
				t.Errorf("JSON marshaling failed: %v", err)
				return
			}

			result := string(jsonBytes)
			if result != tt.expected {
				t.Errorf("JSON marshaling result mismatch:\ngot:  %q\nwant: %q", result, tt.expected)
			}
		})
	}
}
