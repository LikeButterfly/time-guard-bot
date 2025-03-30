package helpers

import (
	"crypto/rand"
	"math/big"
)

// TODO лучше засунуть это в бота
const (
	// Length of generated task IDs
	TaskIDLength = 5

	// Characters used in task IDs
	TaskIDChars = "abcdefghijklmnopqrstuvwxyz0123456789"

	// Maximum number of tasks per group
	MaxTasksPerGroup = 32

	// Maximum number of tasks per user in a group
	MaxTasksPerUser = 4

	// Maximum allowed task duration in minutes (24 hours)
	MaxTaskDuration = 1440

	// Minimum allowed task duration in minutes
	MinTaskDuration = 1
)

// Generates a random task ID of specified length // FIXME!
func GenerateTaskID(length int) (string, error) {
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
