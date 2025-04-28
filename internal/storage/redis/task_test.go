package redis

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"time-guard-bot/internal/models"
)

func TestTaskOperations(t *testing.T) {
	miniRedis, storage := setupMiniRedis(t)
	defer miniRedis.Close()

	ctx := context.Background()
	chatID := int64(12345)
	taskID := "task9"

	// Создаем тестовую задачу
	task := &models.Task{
		ID:          taskID,
		Name:        "Test_Task",
		Description: "Test description",
		ChatID:      chatID,
		OwnerID:     67890,
		StartTime:   time.Now().Truncate(time.Second),
		EndTime:     time.Now().Add(1 * time.Hour).Truncate(time.Second),
		Duration:    60,
		IsLocked:    false,
		LockReason:  "",
		MessageID:   100,
	}

	// Тестируем добавление задачи
	t.Run("AddTask", func(t *testing.T) {
		err := storage.AddTask(ctx, task)
		if err != nil {
			t.Fatalf("Failed to add task: %v", err)
		}

		// Проверяем, что задача добавлена в Redis
		taskKey := fmt.Sprintf(taskIDPrefix, chatID, taskID)

		exists, err := storage.client.Exists(ctx, taskKey).Result()
		if err != nil {
			t.Fatalf("Failed to check if task exists: %v", err)
		}

		if exists != 1 {
			t.Errorf("Task was not added to Redis")
		}

		// Проверяем, что индекс по имени задачи создан
		nameKey := fmt.Sprintf(taskNamePrefix, chatID, task.Name)

		storedID, err := storage.client.Get(ctx, nameKey).Result()
		if err != nil {
			t.Fatalf("Failed to get task ID by name: %v", err)
		}

		if storedID != taskID {
			t.Errorf("Task name index mismatch. got: %s, want: %s", storedID, taskID)
		}
	})

	// Тестируем получение задачи по ID
	t.Run("GetTask", func(t *testing.T) {
		fetchedTask, err := storage.GetTask(ctx, chatID, taskID)
		if err != nil {
			t.Fatalf("Failed to get task: %v", err)
		}

		// Проверяем, что поля задачи соответствуют оригинальным
		if fetchedTask.ID != task.ID {
			t.Errorf("Task ID mismatch. got: %s, want: %s", fetchedTask.ID, task.ID)
		}

		if fetchedTask.Name != task.Name {
			t.Errorf("Task Name mismatch. got: %s, want: %s", fetchedTask.Name, task.Name)
		}

		if fetchedTask.Description != task.Description {
			t.Errorf("Task Description mismatch. got: %s, want: %s", fetchedTask.Description, task.Description)
		}

		if fetchedTask.ChatID != task.ChatID {
			t.Errorf("Task ChatID mismatch. got: %d, want: %d", fetchedTask.ChatID, task.ChatID)
		}

		if fetchedTask.Duration != task.Duration {
			t.Errorf("Task Duration mismatch. got: %d, want: %d", fetchedTask.Duration, task.Duration)
		}

		if !fetchedTask.StartTime.Equal(task.StartTime) {
			t.Errorf("Task StartTime mismatch. got: %v, want: %v", fetchedTask.StartTime, task.StartTime)
		}
	})

	// Тестируем получение несуществующей задачи
	t.Run("GetNonExistentTask", func(t *testing.T) {
		_, err := storage.GetTask(ctx, chatID, "nonex")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound for nonex task, got: %v", err)
		}
	})

	// Тестируем получение задачи по имени
	t.Run("GetTaskByName", func(t *testing.T) {
		fetchedTask, err := storage.GetTaskByName(ctx, chatID, task.Name)
		if err != nil {
			t.Fatalf("Failed to get task by name: %v", err)
		}

		if fetchedTask.ID != task.ID {
			t.Errorf("Task ID mismatch. got: %s, want: %s", fetchedTask.ID, task.ID)
		}
	})

	// Тестируем получение несуществующей задачи по имени
	t.Run("GetNonExistentTaskByName", func(t *testing.T) {
		_, err := storage.GetTaskByName(ctx, chatID, "nonex")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound for nonex task name, got: %v", err)
		}
	})

	// Тестируем обновление задачи
	t.Run("UpdateTask", func(t *testing.T) {
		// Изменяем поля задачи
		updatedTask := &models.Task{
			ID:          taskID,
			Name:        "Updated_Task",
			Description: "Updated description",
			ChatID:      chatID,
			OwnerID:     67890,
			StartTime:   task.StartTime,
			EndTime:     task.EndTime,
			Duration:    120,
			IsLocked:    true,
			LockReason:  "Testing lock",
			MessageID:   101,
		}

		err := storage.UpdateTask(ctx, updatedTask)
		if err != nil {
			t.Fatalf("Failed to update task: %v", err)
		}

		// Проверяем, что старый индекс имени удален
		oldNameKey := fmt.Sprintf(taskNamePrefix, chatID, task.Name)

		exists, err := storage.client.Exists(ctx, oldNameKey).Result()
		if err != nil {
			t.Fatalf("Failed to check if old name index exists: %v", err)
		}

		if exists != 0 {
			t.Errorf("Old task name index was not removed")
		}

		// Проверяем, что новый индекс имени создан
		newNameKey := fmt.Sprintf(taskNamePrefix, chatID, updatedTask.Name)

		storedID, err := storage.client.Get(ctx, newNameKey).Result()
		if err != nil {
			t.Fatalf("Failed to get task ID by new name: %v", err)
		}

		if storedID != taskID {
			t.Errorf("New task name index mismatch. got: %s, want: %s", storedID, taskID)
		}

		// Получаем обновленную задачу и проверяем поля
		fetchedTask, err := storage.GetTask(ctx, chatID, taskID)
		if err != nil {
			t.Fatalf("Failed to get updated task: %v", err)
		}

		if fetchedTask.Name != updatedTask.Name {
			t.Errorf("Updated Task Name mismatch. got: %s, want: %s", fetchedTask.Name, updatedTask.Name)
		}

		if fetchedTask.Description != updatedTask.Description {
			t.Errorf("Updated Task Description mismatch. got: %s, want: %s", fetchedTask.Description, updatedTask.Description)
		}

		if fetchedTask.Duration != updatedTask.Duration {
			t.Errorf("Updated Task Duration mismatch. got: %d, want: %d", fetchedTask.Duration, updatedTask.Duration)
		}

		if fetchedTask.IsLocked != updatedTask.IsLocked {
			t.Errorf("Updated Task IsLocked mismatch. got: %v, want: %v", fetchedTask.IsLocked, updatedTask.IsLocked)
		}

		if fetchedTask.LockReason != updatedTask.LockReason {
			t.Errorf("Updated Task LockReason mismatch. got: %s, want: %s", fetchedTask.LockReason, updatedTask.LockReason)
		}
	})

	// Тестируем обновление несуществующей задачи
	t.Run("UpdateNonExistentTask", func(t *testing.T) {
		nonExistentTask := &models.Task{
			ID:     "nonex",
			ChatID: chatID,
		}

		err := storage.UpdateTask(ctx, nonExistentTask)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound for updating nonex task, got: %v", err)
		}
	})

	// Тестируем получение списка задач
	t.Run("ListTasks", func(t *testing.T) {
		// Добавляем еще одну задачу
		task2 := &models.Task{
			ID:          "tas00",
			Name:        "Second_Task",
			Description: "Another task",
			ChatID:      chatID,
			Duration:    30,
		}

		err := storage.AddTask(ctx, task2)
		if err != nil {
			t.Fatalf("Failed to add second task: %v", err)
		}

		// Получаем список всех задач
		tasks, err := storage.ListTasks(ctx, chatID)
		if err != nil {
			t.Fatalf("Failed to list tasks: %v", err)
		}

		// Проверяем, что получили обе задачи
		if len(tasks) != 2 {
			t.Errorf("Expected 2 tasks, got %d", len(tasks))
		}

		// Проверяем, что задачи содержат правильные ID
		taskIDs := map[string]bool{}
		for _, task := range tasks {
			taskIDs[task.ID] = true
		}

		if !taskIDs[taskID] || !taskIDs["tas00"] {
			t.Errorf("Missing tasks in the list. taskIDs map: %v", taskIDs)
		}
	})

	// Тестируем подсчет задач
	t.Run("CountTasks", func(t *testing.T) {
		count, err := storage.CountTasks(ctx, chatID)
		if err != nil {
			t.Fatalf("Failed to count tasks: %v", err)
		}

		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})

	// Тестируем удаление задачи
	t.Run("DeleteTask", func(t *testing.T) {
		err := storage.DeleteTask(ctx, chatID, taskID)
		if err != nil {
			t.Fatalf("Failed to delete task: %v", err)
		}

		// Проверяем, что задача удалена
		_, err = storage.GetTask(ctx, chatID, taskID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound after deletion, got: %v", err)
		}

		// Проверяем, что индекс имени удален
		nameKey := fmt.Sprintf(taskNamePrefix, chatID, "Updated_Task") // Имя после обновления

		exists, err := storage.client.Exists(ctx, nameKey).Result()
		if err != nil {
			t.Fatalf("Failed to check if name index exists: %v", err)
		}

		if exists != 0 {
			t.Errorf("Task name index was not removed after deletion")
		}

		// Проверяем, что ID задачи удален из списка задач чата
		taskListK := fmt.Sprintf(taskListKey, chatID)

		isMember, err := storage.client.SIsMember(ctx, taskListK, taskID).Result()
		if err != nil {
			t.Fatalf("Failed to check if task ID is still in task list: %v", err)
		}

		if isMember {
			t.Errorf("Task ID was not removed from chat's task list")
		}

		// Проверяем, что вторая задача все еще существует
		count, err := storage.CountTasks(ctx, chatID)
		if err != nil {
			t.Fatalf("Failed to count tasks: %v", err)
		}

		if count != 1 {
			t.Errorf("Expected count 1 after deletion, got %d", count)
		}
	})

	// Тестируем удаление несуществующей задачи
	t.Run("DeleteNonExistentTask", func(t *testing.T) {
		err := storage.DeleteTask(ctx, chatID, "nonex")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound for deleting nonex task, got: %v", err)
		}
	})
}

func TestEmptyTaskList(t *testing.T) {
	miniRedis, storage := setupMiniRedis(t)
	defer miniRedis.Close()

	ctx := context.Background()
	chatID := int64(99999) // Другой чат ID, чтобы не пересекаться с другими тестами

	// Проверяем, что список задач для нового чата пуст
	tasks, err := storage.ListTasks(ctx, chatID)
	if err != nil {
		t.Fatalf("Failed to list tasks for empty chat: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected empty task list for new chat, got %d tasks", len(tasks))
	}

	// Проверяем, что счетчик задач для нового чата равен 0
	count, err := storage.CountTasks(ctx, chatID)
	if err != nil {
		t.Fatalf("Failed to count tasks for empty chat: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected task count 0 for new chat, got %d", count)
	}
}
