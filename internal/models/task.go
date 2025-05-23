package models

import (
	"encoding/json"
	"time"
)

// Represents a task in the system
type Task struct {
	ID          string `json:"id"`          // Short unique identifier for the task
	Name        string `json:"name"`        // Friendly name for the task
	Description string `json:"description"` // Optional description

	ChatID  int64 `json:"chat_id"`  // Telegram chat ID
	OwnerID int64 `json:"owner_id"` // User ID of the person who currently owns the task

	StartTime time.Time `json:"start_time"` // When the task was started
	EndTime   time.Time `json:"end_time"`   // When the task is scheduled to end
	Duration  int       `json:"duration"`   // Duration in minutes

	IsLocked   bool   `json:"is_locked"`   // Whether the task is locked
	LockReason string `json:"lock_reason"` // Reason for locking the task

	MessageID     int `json:"message_id"`      // Original message ID
	BotResponseID int `json:"bot_response_id"` // Bot's response message ID
}

// Represents a task that is currently active
type ActiveTask struct {
	TaskID        string    `json:"task_id"`         // ID of the task
	UserID        int64     `json:"user_id"`         // ID of the user who started the task
	ChatID        int64     `json:"chat_id"`         // ID of the chat where the task was started
	StartTime     time.Time `json:"start_time"`      // When the task was started
	EndTime       time.Time `json:"end_time"`        // When the task is scheduled to end
	Duration      int       `json:"duration"`        // Duration in minutes
	MessageID     int       `json:"message_id"`      // ID of the message in Telegram that started the task
	BotResponseID int       `json:"bot_response_id"` // Bot's response message ID
}

// Marshal converts the task to JSON
func (t *Task) Marshal() ([]byte, error) {
	return json.Marshal(t)
}

// Unmarshal converts JSON to a Task
func UnmarshalTask(data []byte) (*Task, error) {
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

func calcTimeRemaining(startTime time.Time, durationMin int) int64 {
	endTime := startTime.Add(time.Duration(durationMin) * time.Minute)

	remaining := endTime.Unix() - time.Now().Unix()
	if remaining < 0 {
		return 0
	}

	return remaining
}

// Returns the time remaining in seconds
func (t *Task) TimeRemaining() int64 {
	return calcTimeRemaining(t.StartTime, t.Duration)
}

// Returns the time remaining in seconds
func (t *ActiveTask) TimeRemaining() int64 {
	return calcTimeRemaining(t.StartTime, t.Duration)
}
