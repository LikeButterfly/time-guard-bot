package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTaskMarshalUnmarshal(t *testing.T) {
	// Тестовый экземпляр Task
	now := time.Now().Truncate(time.Second)
	task := &Task{
		ID:            "12345",
		Name:          "Test_Task",
		Description:   "Task for testing",
		ChatID:        123456789,
		OwnerID:       987654321,
		StartTime:     now,
		EndTime:       now.Add(2 * time.Hour),
		Duration:      120,
		IsLocked:      false,
		LockReason:    "",
		MessageID:     100,
		BotResponseID: 101,
	}

	// Тестирование Marshal
	data, err := task.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal task: %v", err)
	}

	// Тестирование Unmarshal
	unmarshaledTask, err := UnmarshalTask(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal task: %v", err)
	}

	// Проверка, что данные сохранились корректно
	if task.ID != unmarshaledTask.ID {
		t.Errorf("ID mismatch: expected %s, got %s", task.ID, unmarshaledTask.ID)
	}

	if task.Name != unmarshaledTask.Name {
		t.Errorf("Name mismatch: expected %s, got %s", task.Name, unmarshaledTask.Name)
	}

	if task.ChatID != unmarshaledTask.ChatID {
		t.Errorf("ChatID mismatch: expected %d, got %d", task.ChatID, unmarshaledTask.ChatID)
	}

	if task.OwnerID != unmarshaledTask.OwnerID {
		t.Errorf("OwnerID mismatch: expected %d, got %d", task.OwnerID, unmarshaledTask.OwnerID)
	}

	if !task.StartTime.Equal(unmarshaledTask.StartTime) {
		t.Errorf("StartTime mismatch: expected %v, got %v", task.StartTime, unmarshaledTask.StartTime)
	}

	if !task.EndTime.Equal(unmarshaledTask.EndTime) {
		t.Errorf("EndTime mismatch: expected %v, got %v", task.EndTime, unmarshaledTask.EndTime)
	}

	if task.Duration != unmarshaledTask.Duration {
		t.Errorf("Duration mismatch: expected %d, got %d", task.Duration, unmarshaledTask.Duration)
	}

	if task.IsLocked != unmarshaledTask.IsLocked {
		t.Errorf("IsLocked mismatch: expected %v, got %v", task.IsLocked, unmarshaledTask.IsLocked)
	}

	if task.LockReason != unmarshaledTask.LockReason {
		t.Errorf("LockReason mismatch: expected %s, got %s", task.LockReason, unmarshaledTask.LockReason)
	}

	if task.MessageID != unmarshaledTask.MessageID {
		t.Errorf("MessageID mismatch: expected %d, got %d", task.MessageID, unmarshaledTask.MessageID)
	}

	if task.BotResponseID != unmarshaledTask.BotResponseID {
		t.Errorf("BotResponseID mismatch: expected %d, got %d", task.BotResponseID, unmarshaledTask.BotResponseID)
	}
}

func TestUnmarshalTaskInvalid(t *testing.T) {
	// Тестирование с неправильным JSON
	invalidJSON := []byte(`{"id": "12345", "invalid_json":}`)

	_, err := UnmarshalTask(invalidJSON)
	if err == nil {
		t.Error("Expected error when unmarshaling invalid JSON, but got nil")
	}
}

func TestTaskTimeRemaining(t *testing.T) {
	// Тест для метода TimeRemaining структуры Task
	t.Run("Task with time remaining", func(t *testing.T) {
		now := time.Now()
		task := &Task{
			StartTime: now,
			Duration:  30,
		}

		remaining := task.TimeRemaining()
		// Проверяем, что оставшееся время близко к 30 минутам в секундах (1800)
		if remaining <= 0 || remaining > 1800 {
			t.Errorf("Expected remaining time to be between 0 and 1800 seconds, got %d", remaining)
		}
	})

	t.Run("Task with time expired", func(t *testing.T) {
		pastTime := time.Now().Add(-1 * time.Hour) // 1 час назад
		task := &Task{
			StartTime: pastTime,
			Duration:  30,
		}

		remaining := task.TimeRemaining()
		if remaining != 0 {
			t.Errorf("Expected 0 for expired task, got %d", remaining)
		}
	})

	// Тест граничных случаев
	t.Run("Task with exactly expired time", func(t *testing.T) {
		exactTime := time.Now().Add(-30 * time.Minute) // Ровно 30 минут назад
		task := &Task{
			StartTime: exactTime,
			Duration:  30,
		}

		remaining := task.TimeRemaining()
		if remaining != 0 {
			t.Errorf("Expected 0 for task expired exactly now, got %d", remaining)
		}
	})

	t.Run("Task with very short time remaining", func(t *testing.T) {
		almostExpiredTime := time.Now().Add(-29*time.Minute - 59*time.Second) // Почти 30 минут назад
		task := &Task{
			StartTime: almostExpiredTime,
			Duration:  30,
		}

		remaining := task.TimeRemaining()
		if remaining > 1 {
			t.Errorf("Expected remaining time to be <=1 second, got %d", remaining)
		}
	})
}

func TestActiveTaskTimeRemaining(t *testing.T) {
	// Тест для метода TimeRemaining структуры ActiveTask
	t.Run("ActiveTask with time remaining", func(t *testing.T) {
		now := time.Now()
		activeTask := &ActiveTask{
			StartTime: now,
			Duration:  30,
		}

		remaining := activeTask.TimeRemaining()
		// Проверяем, что оставшееся время близко к 30 минутам в секундах (1800)
		if remaining <= 0 || remaining > 1800 {
			t.Errorf("Expected remaining time to be between 0 and 1800 seconds, got %d", remaining)
		}
	})

	t.Run("ActiveTask with time expired", func(t *testing.T) {
		pastTime := time.Now().Add(-1 * time.Hour) // 1 час назад
		activeTask := &ActiveTask{
			StartTime: pastTime,
			Duration:  30,
		}

		remaining := activeTask.TimeRemaining()
		if remaining != 0 {
			t.Errorf("Expected 0 for expired task, got %d", remaining)
		}
	})

	// Тест граничных случаев
	t.Run("ActiveTask with exactly expired time", func(t *testing.T) {
		exactTime := time.Now().Add(-30 * time.Minute) // Ровно 30 минут назад
		activeTask := &ActiveTask{
			StartTime: exactTime,
			Duration:  30,
		}

		remaining := activeTask.TimeRemaining()
		if remaining != 0 {
			t.Errorf("Expected 0 for task expired exactly now, got %d", remaining)
		}
	})

	t.Run("ActiveTask with very short time remaining", func(t *testing.T) {
		almostExpiredTime := time.Now().Add(-29*time.Minute - 59*time.Second) // Почти 30 минут назад
		activeTask := &ActiveTask{
			StartTime: almostExpiredTime,
			Duration:  30,
		}

		remaining := activeTask.TimeRemaining()
		if remaining > 1 {
			t.Errorf("Expected remaining time to be ≤1 second, got %d", remaining)
		}
	})
}

func TestTaskJSON(t *testing.T) {
	// Проверка, что структура Task корректно маршалится в JSON и обратно
	task := &Task{
		ID:          "12345",
		Name:        "Test_Task",
		Description: "Task for testing",
		ChatID:      123456789,
	}

	// Marshal в JSON вручную для проверки
	jsonBytes, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Failed to marshal task to JSON: %v", err)
	}

	// Unmarshal
	var unmarshaled Task

	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON to task: %v", err)
	}

	// Проверка полей
	if task.ID != unmarshaled.ID || task.Name != unmarshaled.Name ||
		task.Description != unmarshaled.Description || task.ChatID != unmarshaled.ChatID {
		t.Errorf("JSON unmarshaling failed. Original: %+v, Unmarshaled: %+v", task, unmarshaled)
	}
}

func TestActiveTaskJSON(t *testing.T) {
	// Проверка, что структура ActiveTask корректно маршалится в JSON и обратно
	now := time.Now().Truncate(time.Second)
	activeTask := &ActiveTask{
		TaskID:        "12345",
		UserID:        123456789,
		ChatID:        987654321,
		StartTime:     now,
		EndTime:       now.Add(30 * time.Minute),
		Duration:      30,
		MessageID:     100,
		BotResponseID: 101,
	}

	// Marshal в JSON
	jsonBytes, err := json.Marshal(activeTask)
	if err != nil {
		t.Fatalf("Failed to marshal ActiveTask to JSON: %v", err)
	}

	// Unmarshal
	var unmarshaled ActiveTask

	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON to ActiveTask: %v", err)
	}

	// Проверка полей
	if activeTask.TaskID != unmarshaled.TaskID {
		t.Errorf("TaskID mismatch: expected %s, got %s", activeTask.TaskID, unmarshaled.TaskID)
	}

	if activeTask.UserID != unmarshaled.UserID {
		t.Errorf("UserID mismatch: expected %d, got %d", activeTask.UserID, unmarshaled.UserID)
	}

	if activeTask.ChatID != unmarshaled.ChatID {
		t.Errorf("ChatID mismatch: expected %d, got %d", activeTask.ChatID, unmarshaled.ChatID)
	}

	if !activeTask.StartTime.Equal(unmarshaled.StartTime) {
		t.Errorf("StartTime mismatch: expected %v, got %v", activeTask.StartTime, unmarshaled.StartTime)
	}

	if !activeTask.EndTime.Equal(unmarshaled.EndTime) {
		t.Errorf("EndTime mismatch: expected %v, got %v", activeTask.EndTime, unmarshaled.EndTime)
	}

	if activeTask.Duration != unmarshaled.Duration {
		t.Errorf("Duration mismatch: expected %d, got %d", activeTask.Duration, unmarshaled.Duration)
	}

	if activeTask.MessageID != unmarshaled.MessageID {
		t.Errorf("MessageID mismatch: expected %d, got %d", activeTask.MessageID, unmarshaled.MessageID)
	}

	if activeTask.BotResponseID != unmarshaled.BotResponseID {
		t.Errorf("BotResponseID mismatch: expected %d, got %d", activeTask.BotResponseID, unmarshaled.BotResponseID)
	}
}
