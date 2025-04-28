package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"time-guard-bot/internal/models"
	"time-guard-bot/internal/storage/redis"
)

// Mock implementation of storage.Storage interface
type MockStorage struct {
	ChatExistsFunc              func(ctx context.Context, chatID int64) (bool, error)
	GetTaskFunc                 func(ctx context.Context, chatID int64, taskID string) (*models.Task, error)
	GetActiveTaskFunc           func(ctx context.Context, chatID int64, taskID string) (*models.ActiveTask, error)
	ListTasksFunc               func(ctx context.Context, chatID int64) ([]*models.Task, error)
	GetActiveTasksFunc          func(ctx context.Context, chatID int64) ([]*models.ActiveTask, error)
	AddTaskFunc                 func(ctx context.Context, task *models.Task) error
	UpdateTaskFunc              func(ctx context.Context, task *models.Task) error
	GetTaskByNameFunc           func(ctx context.Context, chatID int64, name string) (*models.Task, error)
	DeleteTaskFunc              func(ctx context.Context, chatID int64, taskID string) error
	CountTasksFunc              func(ctx context.Context, chatID int64) (int64, error)
	StartTaskFunc               func(ctx context.Context, activeTask *models.ActiveTask) error
	EndTaskFunc                 func(ctx context.Context, chatID int64, taskID string) error
	GetActiveChatsFunc          func(ctx context.Context) ([]int64, error)
	GetUserActiveTasksFunc      func(ctx context.Context, chatID int64, userID int64) ([]*models.ActiveTask, error)
	GetCountUserActiveTasksFunc func(ctx context.Context, chatID int64, userID int64) (int64, error)
	CloseFunc                   func() error
}

func (m *MockStorage) ChatExists(ctx context.Context, chatID int64) (bool, error) {
	return m.ChatExistsFunc(ctx, chatID)
}

func (m *MockStorage) GetTask(ctx context.Context, chatID int64, taskID string) (*models.Task, error) {
	return m.GetTaskFunc(ctx, chatID, taskID)
}

func (m *MockStorage) GetActiveTask(ctx context.Context, chatID int64, taskID string) (*models.ActiveTask, error) {
	return m.GetActiveTaskFunc(ctx, chatID, taskID)
}

func (m *MockStorage) ListTasks(ctx context.Context, chatID int64) ([]*models.Task, error) {
	return m.ListTasksFunc(ctx, chatID)
}

func (m *MockStorage) GetActiveTasks(ctx context.Context, chatID int64) ([]*models.ActiveTask, error) {
	return m.GetActiveTasksFunc(ctx, chatID)
}

func (m *MockStorage) AddTask(ctx context.Context, task *models.Task) error {
	return m.AddTaskFunc(ctx, task)
}

func (m *MockStorage) UpdateTask(ctx context.Context, task *models.Task) error {
	return m.UpdateTaskFunc(ctx, task)
}

func (m *MockStorage) GetTaskByName(ctx context.Context, chatID int64, name string) (*models.Task, error) {
	return m.GetTaskByNameFunc(ctx, chatID, name)
}

func (m *MockStorage) DeleteTask(ctx context.Context, chatID int64, taskID string) error {
	return m.DeleteTaskFunc(ctx, chatID, taskID)
}

func (m *MockStorage) CountTasks(ctx context.Context, chatID int64) (int64, error) {
	return m.CountTasksFunc(ctx, chatID)
}

func (m *MockStorage) StartTask(ctx context.Context, activeTask *models.ActiveTask) error {
	return m.StartTaskFunc(ctx, activeTask)
}

func (m *MockStorage) EndTask(ctx context.Context, chatID int64, taskID string) error {
	return m.EndTaskFunc(ctx, chatID, taskID)
}

func (m *MockStorage) GetActiveChats(ctx context.Context) ([]int64, error) {
	return m.GetActiveChatsFunc(ctx)
}

func (m *MockStorage) GetUserActiveTasks(ctx context.Context, chatID int64, userID int64) ([]*models.ActiveTask, error) {
	return m.GetUserActiveTasksFunc(ctx, chatID, userID)
}

func (m *MockStorage) GetCountUserActiveTasks(ctx context.Context, chatID int64, userID int64) (int64, error) {
	return m.GetCountUserActiveTasksFunc(ctx, chatID, userID)
}

func (m *MockStorage) Close() error {
	return m.CloseFunc()
}

// Helper function to create a test server with mock storage
func createTestServer() (*Server, *MockStorage) {
	mockStorage := &MockStorage{
		ChatExistsFunc: func(ctx context.Context, chatID int64) (bool, error) {
			return true, nil
		},
		CloseFunc: func() error {
			return nil
		},
	}

	config := &Config{
		Addr: ":8080",
	}

	return NewServer(config, mockStorage), mockStorage
}

func TestAuthMiddleware(t *testing.T) {
	server, mockStorage := createTestServer()

	// Define a simple test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We don't need to use chatID in this test, just check that it exists
		_, ok := GetChatIDFromContext(r.Context())
		if !ok {
			http.Error(w, "Chat ID not found in context", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create a handler with the auth middleware
	handlerWithAuth := server.authMiddleware(testHandler)

	t.Run("Missing Authorization Header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rec := httptest.NewRecorder()

		handlerWithAuth(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}

		var errorResp models.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errorResp); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}

		if errorResp.Error != "Missing Authorization header" {
			t.Errorf("Expected error message '%s', got '%s'", "Missing Authorization header", errorResp.Error)
		}
	})

	t.Run("Invalid Authorization Format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "InvalidFormat")

		rec := httptest.NewRecorder()

		handlerWithAuth(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}

		var errorResp models.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errorResp); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}

		if errorResp.Error != "Invalid Authorization format. Expected: Bearer API_KEY" {
			t.Errorf("Expected error message '%s', got '%s'", "Invalid Authorization format. Expected: Bearer API_KEY", errorResp.Error)
		}
	})

	t.Run("Chat Does Not Exist", func(t *testing.T) {
		// Override the ChatExistsFunc for this test
		mockStorage.ChatExistsFunc = func(ctx context.Context, chatID int64) (bool, error) {
			return false, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer dGc6MTIzNDU=") // Base64 encoded "tg:12345"

		rec := httptest.NewRecorder()

		handlerWithAuth(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
		}

		var errorResp models.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errorResp); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}

		if errorResp.Error != "Chat not found or has no tasks" {
			t.Errorf("Expected error message '%s', got '%s'", "Chat not found or has no tasks", errorResp.Error)
		}
	})

	t.Run("Storage Error", func(t *testing.T) {
		// Override the ChatExistsFunc for this test
		mockStorage.ChatExistsFunc = func(ctx context.Context, chatID int64) (bool, error) {
			return false, errors.New("storage error")
		}

		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer dGc6MTIzNDU=") // Base64 encoded "tg:12345"

		rec := httptest.NewRecorder()

		handlerWithAuth(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}

		var errorResp models.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errorResp); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}

		if errorResp.Error != "Internal server error" {
			t.Errorf("Expected error message '%s', got '%s'", "Internal server error", errorResp.Error)
		}
	})

	t.Run("Successful Authentication", func(t *testing.T) {
		// Reset ChatExistsFunc for success
		mockStorage.ChatExistsFunc = func(ctx context.Context, chatID int64) (bool, error) {
			return true, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer dGc6MTIzNDU=") // Base64 encoded "tg:12345"

		rec := httptest.NewRecorder()

		handlerWithAuth(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		if rec.Body.String() != "OK" {
			t.Errorf("Expected body '%s', got '%s'", "OK", rec.Body.String())
		}
	})
}

func TestHandleTaskStatus(t *testing.T) {
	server, mockStorage := createTestServer()

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/task/status", nil)
		ctx := context.WithValue(req.Context(), ChatIDKey, int64(12345))
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		server.handleTaskStatus(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
		}
	})

	t.Run("Missing ChatID in Context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/task/status", nil)
		rec := httptest.NewRecorder()

		server.handleTaskStatus(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("Missing TaskID Parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/task/status", nil)
		ctx := context.WithValue(req.Context(), ChatIDKey, int64(12345))
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		server.handleTaskStatus(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("Task Not Found", func(t *testing.T) {
		mockStorage.GetTaskFunc = func(ctx context.Context, chatID int64, taskID string) (*models.Task, error) {
			return nil, redis.ErrNotFound
		}

		req := httptest.NewRequest(http.MethodGet, "/api/task/status?task_id=nonexistent", nil)
		ctx := context.WithValue(req.Context(), ChatIDKey, int64(12345))
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		server.handleTaskStatus(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("Storage Error", func(t *testing.T) {
		mockStorage.GetTaskFunc = func(ctx context.Context, chatID int64, taskID string) (*models.Task, error) {
			return nil, errors.New("storage error")
		}

		req := httptest.NewRequest(http.MethodGet, "/api/task/status?task_id=task1", nil)
		ctx := context.WithValue(req.Context(), ChatIDKey, int64(12345))
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		server.handleTaskStatus(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("Task Locked", func(t *testing.T) {
		mockStorage.GetTaskFunc = func(ctx context.Context, chatID int64, taskID string) (*models.Task, error) {
			return &models.Task{
				ID:         "task1",
				Name:       "Task_1",
				IsLocked:   true,
				LockReason: "Under maintenance",
			}, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/api/task/status?task_id=task1", nil)
		ctx := context.WithValue(req.Context(), ChatIDKey, int64(12345))
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		server.handleTaskStatus(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response models.TaskStatusResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}

		if response.Status != "locked" {
			t.Errorf("Expected status '%s', got '%s'", "locked", response.Status)
		}

		if response.LockReason != "Under maintenance" {
			t.Errorf("Expected lock reason '%s', got '%s'", "Under maintenance", response.LockReason)
		}

		if response.TaskName != "Task_1" {
			t.Errorf("Expected task name '%s', got '%s'", "Task_1", response.TaskName)
		}
	})

	t.Run("Task Busy", func(t *testing.T) {
		mockStorage.GetTaskFunc = func(ctx context.Context, chatID int64, taskID string) (*models.Task, error) {
			return &models.Task{
				ID:   "task1",
				Name: "Task_1",
			}, nil
		}
		mockStorage.GetActiveTaskFunc = func(ctx context.Context, chatID int64, taskID string) (*models.ActiveTask, error) {
			return &models.ActiveTask{
				TaskID:    "task1",
				UserID:    12345,
				ChatID:    12345,
				StartTime: time.Now(),
				EndTime:   time.Now().Add(30 * time.Minute),
				Duration:  30,
			}, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/api/task/status?task_id=task1", nil)
		ctx := context.WithValue(req.Context(), ChatIDKey, int64(12345))
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		server.handleTaskStatus(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response models.TaskStatusResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}

		if response.Status != "busy" {
			t.Errorf("Expected status '%s', got '%s'", "busy", response.Status)
		}

		if response.TaskName != "Task_1" {
			t.Errorf("Expected task name '%s', got '%s'", "Task_1", response.TaskName)
		}
	})

	t.Run("Task Free", func(t *testing.T) {
		mockStorage.GetTaskFunc = func(ctx context.Context, chatID int64, taskID string) (*models.Task, error) {
			return &models.Task{
				ID:   "task1",
				Name: "Task_1",
			}, nil
		}
		mockStorage.GetActiveTaskFunc = func(ctx context.Context, chatID int64, taskID string) (*models.ActiveTask, error) {
			return nil, redis.ErrNotFound
		}

		req := httptest.NewRequest(http.MethodGet, "/api/task/status?task_id=task1", nil)
		ctx := context.WithValue(req.Context(), ChatIDKey, int64(12345))
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		server.handleTaskStatus(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response models.TaskStatusResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}

		if response.Status != "free" {
			t.Errorf("Expected status '%s', got '%s'", "free", response.Status)
		}

		if response.TaskName != "Task_1" {
			t.Errorf("Expected task name '%s', got '%s'", "Task_1", response.TaskName)
		}
	})

	t.Run("Error Getting Active Task", func(t *testing.T) {
		mockStorage.GetTaskFunc = func(ctx context.Context, chatID int64, taskID string) (*models.Task, error) {
			return &models.Task{
				ID:   "task1",
				Name: "Task_1",
			}, nil
		}
		mockStorage.GetActiveTaskFunc = func(ctx context.Context, chatID int64, taskID string) (*models.ActiveTask, error) {
			return nil, errors.New("storage error")
		}

		req := httptest.NewRequest(http.MethodGet, "/api/task/status?task_id=task1", nil)
		ctx := context.WithValue(req.Context(), ChatIDKey, int64(12345))
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		server.handleTaskStatus(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})
}

func TestHandleTaskList(t *testing.T) {
	server, mockStorage := createTestServer()

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/task/list", nil)
		ctx := context.WithValue(req.Context(), ChatIDKey, int64(12345))
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		server.handleTaskList(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
		}
	})

	t.Run("Missing ChatID in Context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/task/list", nil)
		rec := httptest.NewRecorder()

		server.handleTaskList(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("Storage Error - ListTasks", func(t *testing.T) {
		mockStorage.ListTasksFunc = func(ctx context.Context, chatID int64) ([]*models.Task, error) {
			return nil, errors.New("storage error")
		}

		req := httptest.NewRequest(http.MethodGet, "/api/task/list", nil)
		ctx := context.WithValue(req.Context(), ChatIDKey, int64(12345))
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		server.handleTaskList(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("Storage Error - GetActiveTasks", func(t *testing.T) {
		mockStorage.ListTasksFunc = func(ctx context.Context, chatID int64) ([]*models.Task, error) {
			return []*models.Task{
				{ID: "task1", Name: "Task_1", Description: "Task_1 description"},
				{ID: "task2", Name: "Task_2", Description: "Task_2 description"},
			}, nil
		}
		mockStorage.GetActiveTasksFunc = func(ctx context.Context, chatID int64) ([]*models.ActiveTask, error) {
			return nil, errors.New("storage error")
		}

		req := httptest.NewRequest(http.MethodGet, "/api/task/list", nil)
		ctx := context.WithValue(req.Context(), ChatIDKey, int64(12345))
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		server.handleTaskList(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("Successful Response - Mixed Statuses", func(t *testing.T) {
		mockStorage.ListTasksFunc = func(ctx context.Context, chatID int64) ([]*models.Task, error) {
			return []*models.Task{
				{ID: "task1", Name: "Task_1", Description: "Task_1 description"},
				{ID: "task2", Name: "Task_2", Description: "Task_2 description", IsLocked: true, LockReason: "Under maintenance"},
				{ID: "task3", Name: "Task_3", Description: "Task_3 description"},
			}, nil
		}
		mockStorage.GetActiveTasksFunc = func(ctx context.Context, chatID int64) ([]*models.ActiveTask, error) {
			return []*models.ActiveTask{
				{TaskID: "task1", UserID: 12345, ChatID: 12345},
			}, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/api/task/list", nil)
		ctx := context.WithValue(req.Context(), ChatIDKey, int64(12345))
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		server.handleTaskList(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response models.TaskListResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}

		// Check task1 is busy
		if task, exists := response["Task_1"]; !exists {
			t.Errorf("Task 'Task_1' not found in response")
		} else if task.Status != "busy" {
			t.Errorf("Expected status '%s' for task 'Task_1', got '%s'", "busy", task.Status)
		}

		// Check task2 is locked
		if task, exists := response["Task_2"]; !exists {
			t.Errorf("Task 'Task_2' not found in response")
		} else if task.Status != "locked" {
			t.Errorf("Expected status '%s' for task 'Task_2', got '%s'", "locked", task.Status)
		} else if task.LockReason != "Under maintenance" {
			t.Errorf("Expected lock reason '%s' for task 'Task_2', got '%s'", "Under maintenance", task.LockReason)
		}

		// Check task3 is free
		if task, exists := response["Task_3"]; !exists {
			t.Errorf("Task 'Task_3' not found in response")
		} else if task.Status != "free" {
			t.Errorf("Expected status '%s' for task 'Task_3', got '%s'", "free", task.Status)
		}
	})

	t.Run("Successful Response - No Tasks", func(t *testing.T) {
		mockStorage.ListTasksFunc = func(ctx context.Context, chatID int64) ([]*models.Task, error) {
			return []*models.Task{}, nil
		}
		mockStorage.GetActiveTasksFunc = func(ctx context.Context, chatID int64) ([]*models.ActiveTask, error) {
			return []*models.ActiveTask{}, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/api/task/list", nil)
		ctx := context.WithValue(req.Context(), ChatIDKey, int64(12345))
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		server.handleTaskList(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response models.TaskListResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}

		if len(response) != 0 {
			t.Errorf("Expected empty response, got %d tasks", len(response))
		}
	})
}
