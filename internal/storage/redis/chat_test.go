// Copyright 2025 LikeButterfly
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
