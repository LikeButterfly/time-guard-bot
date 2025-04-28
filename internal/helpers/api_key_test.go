package helpers

import (
	"encoding/base64"
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestGenerateAPIKey(t *testing.T) {
	testCases := []struct {
		name   string
		chatID int64
	}{
		{name: "Positive ID", chatID: 12345},
		{name: "Negative ID", chatID: -54321},
		{name: "Zero ID", chatID: 0},
		{name: "Large Positive ID", chatID: math.MaxInt64 - 10},
		{name: "Large Negative ID", chatID: math.MinInt64 + 10},
		{name: "Max Int64", chatID: math.MaxInt64},
		{name: "Min Int64", chatID: math.MinInt64},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apiKey := GenerateAPIKey(tc.chatID)

			if apiKey == "" {
				t.Errorf("GenerateAPIKey(%d) returned empty string", tc.chatID)
			}

			_, err := base64.StdEncoding.DecodeString(apiKey)
			if err != nil {
				t.Errorf("GenerateAPIKey(%d) returned invalid base64: %v", tc.chatID, err)
			}

			decoded, _ := base64.StdEncoding.DecodeString(apiKey)
			decodedStr := string(decoded)

			expected := fmt.Sprintf("tg:%d", tc.chatID)
			if decodedStr != expected {
				t.Errorf("GenerateAPIKey(%d) content mismatch. got: %s, want: %s", tc.chatID, decodedStr, expected)
			}
		})
	}
}

func TestAPIKeyRoundTrip(t *testing.T) {
	testCases := []struct {
		name   string
		chatID int64
	}{
		{name: "Positive ID", chatID: 12345},
		{name: "Negative ID", chatID: -54321},
		{name: "Zero ID", chatID: 0},
		{name: "Large Positive ID", chatID: math.MaxInt64 - 10},
		{name: "Large Negative ID", chatID: math.MinInt64 + 10},
		{name: "Max Int64", chatID: math.MaxInt64},
		{name: "Min Int64", chatID: math.MinInt64},
		{name: "Common Telegram ID", chatID: 123456789},
		{name: "Small Negative ID", chatID: -1},
		{name: "Small Positive ID", chatID: 1},
		{name: "Medium Size ID", chatID: 999999999},
		{name: "Boundary Value", chatID: math.MaxInt32},
		{name: "Negative Boundary", chatID: math.MinInt32},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apiKey := GenerateAPIKey(tc.chatID)

			extractedID, err := ExtractChatID(apiKey)
			if err != nil {
				t.Errorf("Failed to extract chatID: %v", err)
			}

			if extractedID != tc.chatID {
				t.Errorf("ChatID mismatch: got %d, want %d", extractedID, tc.chatID)
			}
		})
	}
}

func TestExtractChatID(t *testing.T) {
	testCases := []struct {
		name        string
		apiKey      string
		wantChatID  int64
		wantErr     bool
		errContains string
	}{
		{
			name:       "Valid API key",
			apiKey:     base64.StdEncoding.EncodeToString([]byte("tg:123456")),
			wantChatID: 123456,
			wantErr:    false,
		},
		{
			name:        "Invalid base64",
			apiKey:      "not-base64!@#",
			wantChatID:  0,
			wantErr:     true,
			errContains: "invalid API key format",
		},
		{
			name:        "Missing prefix",
			apiKey:      base64.StdEncoding.EncodeToString([]byte("123456")),
			wantChatID:  0,
			wantErr:     true,
			errContains: "invalid API key format",
		},
		{
			name:        "Invalid chatID",
			apiKey:      base64.StdEncoding.EncodeToString([]byte("tg:not-a-number")),
			wantChatID:  0,
			wantErr:     true,
			errContains: "invalid chat ID in API key",
		},
		{
			name:       "Negative chatID",
			apiKey:     base64.StdEncoding.EncodeToString([]byte("tg:-54321")),
			wantChatID: -54321,
			wantErr:    false,
		},
		{
			name:       "Zero chatID",
			apiKey:     base64.StdEncoding.EncodeToString([]byte("tg:0")),
			wantChatID: 0,
			wantErr:    false,
		},
		{
			name:        "Empty string",
			apiKey:      "",
			wantChatID:  0,
			wantErr:     true,
			errContains: "invalid API key format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chatID, err := ExtractChatID(tc.apiKey)

			if tc.wantErr && err == nil {
				t.Errorf("ExtractChatID(%q) expected error but got nil", tc.apiKey)
			}

			if !tc.wantErr && err != nil {
				t.Errorf("ExtractChatID(%q) unexpected error: %v", tc.apiKey, err)
			}

			if tc.wantErr && err != nil && tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
				t.Errorf("ExtractChatID(%q) error %q does not contain %q", tc.apiKey, err.Error(), tc.errContains)
			}

			if chatID != tc.wantChatID {
				t.Errorf("ExtractChatID(%q) = %d, want %d", tc.apiKey, chatID, tc.wantChatID)
			}
		})
	}
}
