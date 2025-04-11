package redis

import (
	"context"
	"fmt"
)

// Checks if chat has any tasks
func (rs *Storage) ChatExists(ctx context.Context, chatID int64) (bool, error) {
	taskListKey := fmt.Sprintf(taskListKey, chatID)

	exists, err := rs.client.Exists(ctx, taskListKey).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}
