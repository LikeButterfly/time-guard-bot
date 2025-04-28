package redis

import (
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

// настраиваем минисервер Redis для тестирования
func setupMiniRedis(t *testing.T) (*miniredis.Miniredis, *Storage) {
	t.Helper()

	// Создаем мини Redis сервер для тестирования
	miniRedis, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}

	// Создаем клиент Redis с подключением к мини-серверу
	client := redis.NewClient(&redis.Options{
		Addr: miniRedis.Addr(),
	})

	// Создаем хранилище с нашим тестовым клиентом
	storage := &Storage{client: client}

	return miniRedis, storage
}

func TestNewWithInvalidAddress(t *testing.T) {
	// Используем заведомо недействительный адрес
	_, err := New("invalid-address:1234", "", 0)
	if err == nil {
		t.Error("Expected error with invalid Redis address, got nil")
	}
}

func TestCreateConnectionAndClose(t *testing.T) {
	miniRedis, storage := setupMiniRedis(t)
	defer miniRedis.Close()

	if err := storage.Close(); err != nil {
		t.Fatalf("Failed to close Redis connection: %v", err)
	}
}
