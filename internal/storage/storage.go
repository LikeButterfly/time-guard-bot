package storage

import (
	"context"
	"errors"
	"fmt"

	"time-guard-bot/internal/models"
	"time-guard-bot/internal/storage/redis"
)

// Is returned when a requested item is not found
var ErrNotFound = errors.New("not found")

// Storage interface for data storage
type Storage interface {
	// Task operations
	AddTask(ctx context.Context, task *models.Task) error
	GetTask(ctx context.Context, groupID int64, taskID string) (*models.Task, error)
	UpdateTask(ctx context.Context, task *models.Task) error
	GetTaskByName(ctx context.Context, groupID int64, name string) (*models.Task, error)
	DeleteTask(ctx context.Context, groupID int64, taskID string) error
	ListTasks(ctx context.Context, groupID int64) ([]*models.Task, error)
	CountTasks(ctx context.Context, groupID int64) (int64, error)

	// Active Task management
	StartTask(ctx context.Context, activeTask *models.ActiveTask) error
	EndTask(ctx context.Context, groupID int64, taskID string) error
	GetActiveTask(ctx context.Context, groupID int64, taskID string) (*models.ActiveTask, error)
	GetActiveTasks(ctx context.Context, groupID int64) ([]*models.ActiveTask, error)
	GetUserActiveTasks(ctx context.Context, groupID int64, userID int64) ([]*models.ActiveTask, error)
	GetCountUserActiveTasks(ctx context.Context, groupID int64, userID int64) (int64, error)

	// Close connection
	Close() error
}

// Создает новое Redis-хранилище
// Эта функция является фабрикой, которая возвращает реализацию интерфейса Storage
func NewRedisStorage(addr, password string, db int) (Storage, error) {
	// Создаем хранилище Redis из пакета redis
	storage, err := redis.New(addr, password, db)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis storage: %w", err)
	}

	return storage, nil
}
