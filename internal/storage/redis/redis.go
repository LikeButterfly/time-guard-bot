package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Is returned when a requested item is not found
var ErrNotFound = errors.New("not found")

// Redis key prefixes
const (
	// Полная информация о task в JSON формате
	taskIDPrefix = "task_id:%d:%s" // task_id:chatID:taskID
	// Индекс для быстрого обращения к задаче по task_name
	taskNamePrefix = "task_name:%d:%s" // task_name:chatID:taskName
	// list id's всех задач группы
	taskListKey = "tasks:%d" // tasks:chatID
	// Информация об активной задаче
	activeTaskPrefix = "active:%d:%s" // active:chatID:taskID
	// list id's всех активных задач группы
	activeTaskListKey = "active:%d" // active:chatID
	// list id's всех активных задач конкретного пользователя
	userTasksKey = "user:%d:%d" // user:chatID:userID
)

// Implements Storage using Redis
type Storage struct {
	client *redis.Client
}

// Creates a new Redis storage
func New(addr, password string, db int) (*Storage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Storage{client: client}, nil
}

// Closes the Redis connection
func (rs *Storage) Close() error {
	return rs.client.Close()
}
