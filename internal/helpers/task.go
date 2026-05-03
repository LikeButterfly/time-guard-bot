package helpers

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// TODO лучше засунуть это в бота
const (
	// Length of generated task IDs
	TaskIDLength = 5

	// Characters used in task IDs
	TaskIDChars = "abcdefghijklmnopqrstuvwxyz0123456789"

	// Maximum number of tasks per chat
	MaxTasksPerChat = 16

	// Maximum number of tasks per user in a chat
	MaxTasksPerUser = 4

	// Maximum allowed task duration in minutes (24 hours)
	MaxTaskDuration = 1440

	// Minimum allowed task duration in minutes
	MinTaskDuration = 1
)

// Generates a random task ID of specified length with uniqueness check
func GenerateUniqueTaskID(length int, existsFunc func(string) (bool, error)) (string, error) {
	const maxRetries = 10

	for attempt := 0; attempt < maxRetries; attempt++ {
		id, err := generateRandomID(length)
		if err != nil {
			return "", fmt.Errorf("failed to generate ID: %w", err)
		}

		// Check if ID already exists
		exists, err := existsFunc(id)
		if err != nil {
			return "", fmt.Errorf("failed to check ID uniqueness: %w", err)
		}

		if !exists {
			return id, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique ID after %d attempts", maxRetries)
}

// Generates a random task ID of specified length (without uniqueness check)
func generateRandomID(length int) (string, error) {
	id := make([]byte, length)
	charCount := big.NewInt(int64(len(TaskIDChars)))

	for i := range length {
		idx, err := rand.Int(rand.Reader, charCount)
		if err != nil {
			return "", err
		}

		id[i] = TaskIDChars[idx.Int64()]
	}

	return string(id), nil
}
