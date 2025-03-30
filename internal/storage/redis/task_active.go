package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"time-guard-bot/internal/models"
)

// Starts a task
func (rs *Storage) StartTask(ctx context.Context, activeTask *models.ActiveTask) error {
	// Check if task exists
	task, err := rs.GetTask(ctx, activeTask.GroupID, activeTask.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// TODO do with Storage method?
	// Check if task is already active
	activeTaskKey := fmt.Sprintf(activeTaskPrefix, activeTask.GroupID, activeTask.TaskID)
	exists, err := rs.client.Exists(ctx, activeTaskKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check if task is active: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("task is already active")
	}

	// Check if task is locked
	if task.IsLocked {
		return fmt.Errorf("task is locked: %s", task.LockReason) // FIXME?
	}

	// Count user's active tasks
	userTasksKey := fmt.Sprintf(userTasksKey, activeTask.GroupID, activeTask.UserID) // FIXME

	// Marshal active task to JSON
	activeTaskJSON, err := json.Marshal(activeTask)
	if err != nil {
		return fmt.Errorf("failed to marshal active task: %w", err)
	}

	// Update task status
	task.OwnerID = activeTask.UserID
	task.StartTime = activeTask.StartTime
	task.EndTime = activeTask.EndTime
	task.Duration = activeTask.Duration
	task.MessageID = activeTask.MessageID
	task.BotResponseID = activeTask.BotResponseID

	// Marshal updated task to JSON
	taskJSON, err := task.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Get Redis pipeline
	pipe := rs.client.Pipeline()

	// Save active task
	pipe.Set(ctx, activeTaskKey, activeTaskJSON, 0)

	// Add to active task list
	activeTaskListKey := fmt.Sprintf(activeTaskListKey, activeTask.GroupID)
	pipe.SAdd(ctx, activeTaskListKey, activeTask.TaskID)

	// Add to user's active tasks
	pipe.SAdd(ctx, userTasksKey, activeTask.TaskID)

	// Update task status
	taskKey := fmt.Sprintf(taskIDPrefix, task.GroupID, task.ID)
	pipe.Set(ctx, taskKey, taskJSON, 0)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to start task: %w", err)
	}

	return nil
}

// Ends a task
func (rs *Storage) EndTask(ctx context.Context, groupID int64, taskID string) error {
	// Get active task
	activeTask, err := rs.GetActiveTask(ctx, groupID, taskID)
	if err != nil {
		return fmt.Errorf("failed to get active task: %w", err)
	}

	// Get task
	task, err := rs.GetTask(ctx, groupID, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Update task status
	task.OwnerID = 0
	task.StartTime = time.Time{}
	task.EndTime = time.Time{}
	task.Duration = 0
	task.MessageID = 0

	// Marshal updated task to JSON
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Get Redis pipeline
	pipe := rs.client.Pipeline()

	// Remove active task
	activeTaskKey := fmt.Sprintf(activeTaskPrefix, groupID, taskID)
	pipe.Del(ctx, activeTaskKey)

	// Remove from active task list
	activeTaskListKey := fmt.Sprintf(activeTaskListKey, groupID)
	pipe.SRem(ctx, activeTaskListKey, taskID)

	// Remove from user's active tasks
	userTasksKey := fmt.Sprintf(userTasksKey, groupID, activeTask.UserID)
	pipe.SRem(ctx, userTasksKey, taskID)

	// Update task status
	taskKey := fmt.Sprintf(taskIDPrefix, task.GroupID, task.ID)
	pipe.Set(ctx, taskKey, taskJSON, 0)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to end task: %w", err)
	}

	return nil
}

// Gets an active task by ID
func (rs *Storage) GetActiveTask(ctx context.Context, groupID int64, taskID string) (*models.ActiveTask, error) {
	activeTaskKey := fmt.Sprintf(activeTaskPrefix, groupID, taskID)
	activeTaskJSON, err := rs.client.Get(ctx, activeTaskKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get active task: %w", err)
	}

	var activeTask models.ActiveTask
	err = json.Unmarshal([]byte(activeTaskJSON), &activeTask)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal active task: %w", err)
	}

	return &activeTask, nil
}

// Gets all active tasks for a group
func (rs *Storage) GetActiveTasks(ctx context.Context, groupID int64) ([]*models.ActiveTask, error) {
	activeTaskListKey := fmt.Sprintf(activeTaskListKey, groupID)
	taskIDs, err := rs.client.SMembers(ctx, activeTaskListKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get active tasks: %w", err)
	}

	activeTasks := make([]*models.ActiveTask, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		activeTask, err := rs.GetActiveTask(ctx, groupID, taskID)
		if err != nil {
			if err == ErrNotFound {
				// Skip task if not found (could happen if task was ended in another goroutine)
				continue
			}
			return nil, fmt.Errorf("failed to get active task: %w", err)
		}
		activeTasks = append(activeTasks, activeTask)
	}

	return activeTasks, nil
}

// Gets all active tasks for a user
func (rs *Storage) GetUserActiveTasks(ctx context.Context, groupID int64, userID int64) ([]*models.ActiveTask, error) {
	userTasksKey := fmt.Sprintf(userTasksKey, groupID, userID)
	taskIDs, err := rs.client.SMembers(ctx, userTasksKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user's active tasks: %w", err)
	}

	activeTasks := make([]*models.ActiveTask, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		activeTask, err := rs.GetActiveTask(ctx, groupID, taskID)
		if err != nil {
			if err == ErrNotFound {
				// Skip task if not found (could happen if task was ended in another goroutine)
				continue
			}
			return nil, fmt.Errorf("failed to get active task: %w", err)
		}
		activeTasks = append(activeTasks, activeTask)
	}

	return activeTasks, nil
}

// Counts the number of active timers for a user in a group
func (rs *Storage) GetCountUserActiveTasks(ctx context.Context, groupID int64, userID int64) (int64, error) {
	userTasksK := fmt.Sprintf(userTasksKey, groupID, userID)
	count, err := rs.client.SCard(ctx, userTasksK).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to count user timers: %w", err)
	}

	return count, nil
}
