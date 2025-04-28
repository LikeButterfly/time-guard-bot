package redis

import (
	"context"
	"testing"

	"time-guard-bot/internal/models"
)

func TestChatOperations(t *testing.T) {
	miniRedis, storage := setupMiniRedis(t)
	defer miniRedis.Close()

	ctx := context.Background()
	chatID := int64(12345)

	t.Run("ChatDoesNotExistInitially", func(t *testing.T) {
		exists, err := storage.ChatExists(ctx, chatID)
		if err != nil {
			t.Fatalf("Failed to check if chat exists: %v", err)
		}

		if exists {
			t.Errorf("Expected chat to not exist initially")
		}
	})

	t.Run("ChatExistsAfterAddingTask", func(t *testing.T) {
		task := &models.Task{
			ID:     "task9",
			ChatID: chatID,
			Name:   "Test_Task",
		}

		err := storage.AddTask(ctx, task)
		if err != nil {
			t.Fatalf("Failed to add task: %v", err)
		}

		exists, err := storage.ChatExists(ctx, chatID)
		if err != nil {
			t.Fatalf("Failed to check if chat exists: %v", err)
		}

		if !exists {
			t.Errorf("Expected chat to exist after adding a task")
		}
	})
}
