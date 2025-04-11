package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

// Extracts the chat ID from an API key
func ExtractChatID(ctx context.Context, key string) (int64, error) {
	// Decoding Base64-encoded key
	decodedBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return 0, fmt.Errorf("invalid API key format: %w", err)
	}

	// Bytes to a string
	decodedStr := string(decodedBytes)

	// Check the format (must be "tg:<chatID>")
	parts := strings.Split(decodedStr, ":")
	if len(parts) != 2 || parts[0] != "tg" {
		return 0, fmt.Errorf("invalid API key format")
	}

	// Extracting chatID
	chatID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid chat ID in API key: %w", err)
	}

	return chatID, nil
}
