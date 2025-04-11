package helpers

import (
	"encoding/base64"
	"fmt"
)

// Generates API key based on chatID
func GenerateAPIKey(chatID int64) string {
	keyData := fmt.Sprintf("tg:%d", chatID)
	key := base64.StdEncoding.EncodeToString([]byte(keyData))

	return key
}
