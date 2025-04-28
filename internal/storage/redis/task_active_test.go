package redis

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"time-guard-bot/internal/models"
)

func TestActiveTaskOperations(t *testing.T) {
	miniRedis, storage := setupMiniRedis(t)
	defer miniRedis.Close()

	ctx := context.Background()
	chatID := int64(12345)
	userID := int64(67890)
	taskID := "task9"

	// Сначала создаем обычную задачу для активации
	baseTask := &models.Task{
		ID:          taskID,
		Name:        "Test_Task",
		Description: "Test description",
		ChatID:      chatID,
		OwnerID:     userID,
		Duration:    60,
	}

	err := storage.AddTask(ctx, baseTask)
	if err != nil {
		t.Fatalf("Failed to add base task: %v", err)
	}

	// Создаем активную задачу
	now := time.Now().Truncate(time.Second)
	activeTask := &models.ActiveTask{
		TaskID:        taskID,
		UserID:        userID,
		ChatID:        chatID,
		StartTime:     now,
		EndTime:       now.Add(60 * time.Minute),
		Duration:      60,
		MessageID:     100,
		BotResponseID: 101,
	}

	// Тестируем добавление активной задачи
	t.Run("StartTask", func(t *testing.T) {
		err := storage.StartTask(ctx, activeTask)
		if err != nil {
			t.Fatalf("Failed to start task: %v", err)
		}

		// Проверяем, что активная задача добавлена в Redis
		activeKey := fmt.Sprintf(activeTaskPrefix, chatID, taskID)

		exists, err := storage.client.Exists(ctx, activeKey).Result()
		if err != nil {
			t.Fatalf("Failed to check if active task exists: %v", err)
		}

		if exists != 1 {
			t.Errorf("Active task was not added to Redis")
		}

		// Проверяем, что задача добавлена в список активных задач чата
		activeListKey := fmt.Sprintf(activeTaskListKey, chatID)

		isMember, err := storage.client.SIsMember(ctx, activeListKey, taskID).Result()
		if err != nil {
			t.Fatalf("Failed to check if task is in active list: %v", err)
		}

		if !isMember {
			t.Errorf("Task was not added to chat's active task list")
		}

		// Проверяем, что задача добавлена в список активных задач пользователя
		userKey := fmt.Sprintf(userTasksKey, chatID, userID)

		isMember, err = storage.client.SIsMember(ctx, userKey, taskID).Result()
		if err != nil {
			t.Fatalf("Failed to check if task is in user's task list: %v", err)
		}

		if !isMember {
			t.Errorf("Task was not added to user's active task list")
		}
	})

	// Тестируем получение активной задачи
	t.Run("GetActiveTask", func(t *testing.T) {
		fetchedTask, err := storage.GetActiveTask(ctx, chatID, taskID)
		if err != nil {
			t.Fatalf("Failed to get active task: %v", err)
		}

		// Проверяем, что поля активной задачи соответствуют оригинальным
		if fetchedTask.TaskID != activeTask.TaskID {
			t.Errorf("TaskID mismatch. got: %s, want: %s", fetchedTask.TaskID, activeTask.TaskID)
		}

		if fetchedTask.UserID != activeTask.UserID {
			t.Errorf("UserID mismatch. got: %d, want: %d", fetchedTask.UserID, activeTask.UserID)
		}

		if fetchedTask.ChatID != activeTask.ChatID {
			t.Errorf("ChatID mismatch. got: %d, want: %d", fetchedTask.ChatID, activeTask.ChatID)
		}

		if fetchedTask.Duration != activeTask.Duration {
			t.Errorf("Duration mismatch. got: %d, want: %d", fetchedTask.Duration, activeTask.Duration)
		}

		if !fetchedTask.StartTime.Equal(activeTask.StartTime) {
			t.Errorf("StartTime mismatch. got: %v, want: %v", fetchedTask.StartTime, activeTask.StartTime)
		}

		if !fetchedTask.EndTime.Equal(activeTask.EndTime) {
			t.Errorf("EndTime mismatch. got: %v, want: %v", fetchedTask.EndTime, activeTask.EndTime)
		}

		if fetchedTask.MessageID != activeTask.MessageID {
			t.Errorf("MessageID mismatch. got: %d, want: %d", fetchedTask.MessageID, activeTask.MessageID)
		}

		if fetchedTask.BotResponseID != activeTask.BotResponseID {
			t.Errorf("BotResponseID mismatch. got: %d, want: %d", fetchedTask.BotResponseID, activeTask.BotResponseID)
		}
	})

	// Тестируем получение несуществующей активной задачи
	t.Run("GetNonExistentActiveTask", func(t *testing.T) {
		_, err := storage.GetActiveTask(ctx, chatID, "nonex")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound for nonex active task, got: %v", err)
		}
	})

	// Тестируем получение списка активных задач чата
	t.Run("GetActiveTasks", func(t *testing.T) {
		// Добавляем еще одну активную задачу
		taskID2 := "tas00"
		baseTask2 := &models.Task{
			ID:          taskID2,
			Name:        "Second_Task",
			Description: "Another task",
			ChatID:      chatID,
			OwnerID:     userID,
			Duration:    30,
		}

		err := storage.AddTask(ctx, baseTask2)
		if err != nil {
			t.Fatalf("Failed to add second base task: %v", err)
		}

		activeTask2 := &models.ActiveTask{
			TaskID:    taskID2,
			UserID:    userID,
			ChatID:    chatID,
			StartTime: now,
			EndTime:   now.Add(30 * time.Minute),
			Duration:  30,
		}

		err = storage.StartTask(ctx, activeTask2)
		if err != nil {
			t.Fatalf("Failed to start second task: %v", err)
		}

		// Получаем список всех активных задач
		activeTasks, err := storage.GetActiveTasks(ctx, chatID)
		if err != nil {
			t.Fatalf("Failed to get active tasks: %v", err)
		}

		// Проверяем, что получили обе активные задачи
		if len(activeTasks) != 2 {
			t.Errorf("Expected 2 active tasks, got %d", len(activeTasks))
		}

		// Проверяем, что задачи содержат правильные ID
		taskIDs := map[string]bool{}
		for _, task := range activeTasks {
			taskIDs[task.TaskID] = true
		}

		if !taskIDs[taskID] || !taskIDs[taskID2] {
			t.Errorf("Missing tasks in the active list. taskIDs map: %v", taskIDs)
		}
	})

	// Тестируем получение списка активных задач пользователя
	t.Run("GetUserActiveTasks", func(t *testing.T) {
		userTasks, err := storage.GetUserActiveTasks(ctx, chatID, userID)
		if err != nil {
			t.Fatalf("Failed to get user active tasks: %v", err)
		}

		// Проверяем, что получили две активные задачи пользователя
		if len(userTasks) != 2 {
			t.Errorf("Expected 2 user active tasks, got %d", len(userTasks))
		}
	})

	// Тестируем подсчет активных задач пользователя
	t.Run("GetCountUserActiveTasks", func(t *testing.T) {
		count, err := storage.GetCountUserActiveTasks(ctx, chatID, userID)
		if err != nil {
			t.Fatalf("Failed to count user active tasks: %v", err)
		}

		if count != 2 {
			t.Errorf("Expected count 2 for user active tasks, got %d", count)
		}
	})

	// Тестируем получение списка чатов с активными задачами
	t.Run("GetActiveChats", func(t *testing.T) {
		// Добавляем активную задачу в другой чат
		otherChatID := int64(54321)
		otherTaskID := "other"

		otherBaseTask := &models.Task{
			ID:          otherTaskID,
			Name:        "Other_Task",
			Description: "Task in other chat",
			ChatID:      otherChatID,
			OwnerID:     userID,
			Duration:    45,
		}

		err := storage.AddTask(ctx, otherBaseTask)
		if err != nil {
			t.Fatalf("Failed to add task in other chat: %v", err)
		}

		otherActiveTask := &models.ActiveTask{
			TaskID:    otherTaskID,
			UserID:    userID,
			ChatID:    otherChatID,
			StartTime: now,
			EndTime:   now.Add(45 * time.Minute),
			Duration:  45,
		}

		err = storage.StartTask(ctx, otherActiveTask)
		if err != nil {
			t.Fatalf("Failed to start task in other chat: %v", err)
		}

		// Получаем список всех чатов с активными задачами
		activeChats, err := storage.GetActiveChats(ctx)
		if err != nil {
			t.Fatalf("Failed to get active chats: %v", err)
		}

		// Проверяем, что получили оба чата
		if len(activeChats) != 2 {
			t.Errorf("Expected 2 active chats, got %d", len(activeChats))
		}

		// Проверяем, что чаты содержат правильные ID
		chatIDs := map[int64]bool{}
		for _, chat := range activeChats {
			chatIDs[chat] = true
		}

		if !chatIDs[chatID] || !chatIDs[otherChatID] {
			t.Errorf("Missing chats in the active list. chatIDs map: %v", chatIDs)
		}
	})

	// Тестируем завершение активной задачи
	t.Run("EndTask", func(t *testing.T) {
		err := storage.EndTask(ctx, chatID, taskID)
		if err != nil {
			t.Fatalf("Failed to end task: %v", err)
		}

		// Проверяем, что активная задача удалена из Redis
		activeKey := fmt.Sprintf(activeTaskPrefix, chatID, taskID)

		exists, err := storage.client.Exists(ctx, activeKey).Result()
		if err != nil {
			t.Fatalf("Failed to check if active task exists: %v", err)
		}

		if exists != 0 {
			t.Errorf("Active task was not removed from Redis")
		}

		// Проверяем, что задача удалена из списка активных задач чата
		activeListKey := fmt.Sprintf(activeTaskListKey, chatID)

		isMember, err := storage.client.SIsMember(ctx, activeListKey, taskID).Result()
		if err != nil {
			t.Fatalf("Failed to check if task is in active list: %v", err)
		}

		if isMember {
			t.Errorf("Task was not removed from chat's active task list")
		}

		// Проверяем, что задача удалена из списка активных задач пользователя
		userKey := fmt.Sprintf(userTasksKey, chatID, userID)

		isMember, err = storage.client.SIsMember(ctx, userKey, taskID).Result()
		if err != nil {
			t.Fatalf("Failed to check if task is in user's task list: %v", err)
		}

		if isMember {
			t.Errorf("Task was not removed from user's active task list")
		}

		// Проверяем, что оставшиеся задачи все еще существуют
		remainingTasks, err := storage.GetActiveTasks(ctx, chatID)
		if err != nil {
			t.Fatalf("Failed to get remaining active tasks: %v", err)
		}

		if len(remainingTasks) != 1 {
			t.Errorf("Expected 1 remaining active task, got %d", len(remainingTasks))
		}
	})

	// Тестируем завершение несуществующей активной задачи
	t.Run("EndNonExistentTask", func(t *testing.T) {
		err := storage.EndTask(ctx, chatID, "nonex")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound for ending nonex task, got: %v", err)
		}
	})
}

func TestEmptyActiveTasks(t *testing.T) {
	miniRedis, storage := setupMiniRedis(t)
	defer miniRedis.Close()

	ctx := context.Background()
	chatID := int64(99999) // Другой чат ID
	userID := int64(88888) // Другой пользователь ID

	// Проверяем, что список активных задач для нового чата пуст
	activeTasks, err := storage.GetActiveTasks(ctx, chatID)
	if err != nil {
		t.Fatalf("Failed to get active tasks for empty chat: %v", err)
	}

	if len(activeTasks) != 0 {
		t.Errorf("Expected empty active tasks list for new chat, got %d tasks", len(activeTasks))
	}

	// Проверяем, что список активных задач пользователя пуст
	userTasks, err := storage.GetUserActiveTasks(ctx, chatID, userID)
	if err != nil {
		t.Fatalf("Failed to get user active tasks for empty chat: %v", err)
	}

	if len(userTasks) != 0 {
		t.Errorf("Expected empty user active tasks list, got %d tasks", len(userTasks))
	}

	// Проверяем, что счетчик активных задач пользователя равен 0
	count, err := storage.GetCountUserActiveTasks(ctx, chatID, userID)
	if err != nil {
		t.Fatalf("Failed to count user active tasks for empty chat: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected user active tasks count 0, got %d", count)
	}
}
