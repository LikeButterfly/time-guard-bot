package redis

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"

	"time-guard-bot/internal/models"
)

// Добавляет новый task
func (rs *Storage) AddTask(ctx context.Context, task *models.Task) error {
	// TODO do here gen task id
	// мб тут и проверку, сущ-ет ли задача с таким же Name
	// Marshal task to JSON
	taskData, err := task.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Create transaction
	pipe := rs.client.TxPipeline()

	// Store task by ID
	taskIDKey := fmt.Sprintf(taskIDPrefix, task.GroupID, task.ID)
	pipe.Set(ctx, taskIDKey, taskData, 0)

	// Create index by name for quick lookup
	if task.Name != "" {
		taskShortKey := fmt.Sprintf(taskNamePrefix, task.GroupID, task.Name)
		pipe.Set(ctx, taskShortKey, task.ID, 0)
	}

	// Add to group's task list
	taskListK := fmt.Sprintf(taskListKey, task.GroupID)
	pipe.SAdd(ctx, taskListK, task.ID)

	// Execute transaction
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to add task: %w", err)
	}

	return nil
}

// Извлекает task по id
func (rs *Storage) GetTask(ctx context.Context, groupID int64, taskID string) (*models.Task, error) {
	taskKey := fmt.Sprintf(taskIDPrefix, groupID, taskID)

	data, err := rs.client.Get(ctx, taskKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	task, err := models.UnmarshalTask(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return task, nil
}

// Обновляет существующий task
func (rs *Storage) UpdateTask(ctx context.Context, task *models.Task) error {
	// Check if task exists
	key := fmt.Sprintf(taskIDPrefix, task.GroupID, task.ID)

	exists, err := rs.client.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to check if task exists: %w", err)
	}

	if exists == 0 {
		return ErrNotFound
	}

	// Get existing task to check if name changed
	existingData, err := rs.client.Get(ctx, key).Bytes()
	if err != nil {
		return fmt.Errorf("failed to get existing task: %w", err)
	}

	existingTask, err := models.UnmarshalTask(existingData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal existing task: %w", err)
	}

	// Marshal updated task
	taskData, err := task.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Create transaction
	pipe := rs.client.TxPipeline()

	// Update task data
	pipe.Set(ctx, key, taskData, 0)

	// Update name index if name changed
	if existingTask.Name != task.Name {
		// Remove old index
		if existingTask.Name != "" {
			oldShortKey := fmt.Sprintf(taskNamePrefix, task.GroupID, existingTask.Name)
			pipe.Del(ctx, oldShortKey)
		}

		// Add new index
		if task.Name != "" {
			newShortKey := fmt.Sprintf(taskNamePrefix, task.GroupID, task.Name)
			pipe.Set(ctx, newShortKey, task.ID, 0)
		}
	}

	// Execute transaction
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

// Извлекает task по name
func (rs *Storage) GetTaskByName(ctx context.Context, groupID int64, name string) (*models.Task, error) {
	// Get task ID from name index
	nameKey := fmt.Sprintf(taskNamePrefix, groupID, name)

	taskIDResult, err := rs.client.Get(ctx, nameKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to get task ID by name: %w", err)
	}

	// Get task by ID
	return rs.GetTask(ctx, groupID, taskIDResult)
}

// Удаляет task по id
func (rs *Storage) DeleteTask(ctx context.Context, groupID int64, taskID string) error {
	// Get task to retrieve name for index deletion
	task, err := rs.GetTask(ctx, groupID, taskID)
	if err != nil {
		return err
	}

	// Create transaction
	pipe := rs.client.TxPipeline()

	// Delete task data
	key := fmt.Sprintf(taskIDPrefix, groupID, taskID)
	pipe.Del(ctx, key)

	// Delete name index
	if task.Name != "" {
		shortKey := fmt.Sprintf(taskNamePrefix, groupID, task.Name)
		pipe.Del(ctx, shortKey)
	}

	// Remove from task list
	taskListK := fmt.Sprintf(taskListKey, groupID)
	pipe.SRem(ctx, taskListK, taskID)

	// Execute transaction
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}

// Извлекает все tasks группы
func (rs *Storage) ListTasks(ctx context.Context, groupID int64) ([]*models.Task, error) {
	// Get all task IDs for the group
	taskListK := fmt.Sprintf(taskListKey, groupID)

	taskIDs, err := rs.client.SMembers(ctx, taskListK).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get task IDs: %w", err)
	}

	// No tasks found
	if len(taskIDs) == 0 {
		return []*models.Task{}, nil
	}

	// Get each task
	tasks := make([]*models.Task, 0, len(taskIDs))

	for _, id := range taskIDs {
		task, err := rs.GetTask(ctx, groupID, id)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				// Skip not found tasks (should not happen in normal operation)
				continue
			}

			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// Считает кол-во tasks группы
func (rs *Storage) CountTasks(ctx context.Context, groupID int64) (int64, error) {
	taskListK := fmt.Sprintf(taskListKey, groupID)

	count, err := rs.client.SCard(ctx, taskListK).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	return count, nil
}
