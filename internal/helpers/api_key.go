package helpers

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

// Generates API key based on chatID
func GenerateAPIKey(chatID int64) string {
	keyData := fmt.Sprintf("tg:%d", chatID)
	key := base64.StdEncoding.EncodeToString([]byte(keyData))

	return key
}

// Extracts the chat ID from an API key
func ExtractChatID(apiKey string) (int64, error) {
	// Decoding Base64-encoded key
	decodedBytes, err := base64.StdEncoding.DecodeString(apiKey)
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
