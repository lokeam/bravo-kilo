package handlers

import (
	"testing"

	"github.com/microcosm-cc/bluemonday"
)

func TestSanitizeAndUnescape(t *testing.T) {
	h := Handlers{
		sanitizer: bluemonday.StrictPolicy(),
	}

	// Define test cases
	testCases := []struct {
		input    string
		expected string
	}{
		// Test sanitization (removes unsafe HTML)
		{
			input:    "<script>alert('attack')</script>Safe text",
			expected: "Safe text",
		},
		// Test unescaping (converts &#39; to ')
		{
			input:    "Can&#39;t Hurt Me",
			expected: "Can't Hurt Me",
		},
		// Test a combination of both
		{
			input:    "<b>Bold Text&#39;s Test</b>",
			expected: "Bold Text's Test",
		},
		// Test plain text
		{
			input:    "Some plain text",
			expected: "Some plain text",
		},
	}

	for _, tc := range testCases {
		result := h.sanitizeAndUnescape(tc.input)
		if result != tc.expected {
			t.Errorf("sanitizeAndUnescape(%q) = %q; want %q", tc.input, result, tc.expected)
		}
	}
}
