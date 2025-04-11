package api

import (
	"log"
	"net/http"

	"time-guard-bot/internal/models"
	"time-guard-bot/internal/storage/redis"
)

// Handles the task status endpoint
func (s *Server) handleTaskStatus(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get chatID from context
	chatID, ok := GetChatIDFromContext(r.Context())
	if !ok {
		sendJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get task ID from query parameters
	taskID := r.URL.Query().Get("task_id")

	// Validate parameters
	if taskID == "" {
		sendJSONError(w, "Missing required parameter: task_id", http.StatusBadRequest)
		return
	}

	// Get task from storage
	task, err := s.storage.GetTask(r.Context(), chatID, taskID)
	if err != nil {
		if err == redis.ErrNotFound {
			sendJSONError(w, "Task not found", http.StatusNotFound)
		} else {
			log.Printf("Failed to get task: %v", err)
			sendJSONError(w, "Internal server error", http.StatusInternalServerError)
		}

		return
	}

	response := models.TaskStatusResponse{
		TaskName: task.Name,
	}

	if task.IsLocked {
		response.Status = "locked"
		response.LockReason = task.LockReason
		sendJSON(w, response)

		return
	}

	activeTask, err := s.storage.GetActiveTask(r.Context(), chatID, taskID)
	if err != nil {
		if err != redis.ErrNotFound {
			log.Printf("Failed to get active task: %v", err)
			sendJSONError(w, "Internal server error", http.StatusInternalServerError)

			return
		}
	}

	if activeTask != nil {
		response.Status = "busy"
		sendJSON(w, response)

		return
	}

	response.Status = "free"
	sendJSON(w, response)
}

// Handles the task list endpoint - returns all chat tasks
func (s *Server) handleTaskList(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get chatID from context
	chatID, ok := GetChatIDFromContext(r.Context())
	if !ok {
		sendJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get all tasks for the chat
	tasks, err := s.storage.ListTasks(r.Context(), chatID)
	if err != nil {
		log.Printf("Error listing tasks: %v", err)
		sendJSONError(w, "Internal server error", http.StatusInternalServerError)

		return
	}

	// Get all active tasks for the chat
	activeTasks, err := s.storage.GetActiveTasks(r.Context(), chatID)
	if err != nil {
		log.Printf("Failed to get active tasks: %v", err)
		sendJSONError(w, "Internal server error", http.StatusInternalServerError)

		return
	}

	// Create a map of active task IDs for quick lookup
	activeTaskMap := make(map[string]bool)
	for _, activeTask := range activeTasks {
		activeTaskMap[activeTask.TaskID] = true
	}

	// Build the response
	response := make(models.TaskListResponse)

	for _, task := range tasks {
		taskInfo := models.TaskInfo{
			ID:          task.ID,
			Description: task.Description,
		}

		if task.IsLocked {
			taskInfo.Status = "locked"
			taskInfo.LockReason = task.LockReason
		} else if activeTaskMap[task.ID] {
			taskInfo.Status = "busy"
		} else {
			taskInfo.Status = "free"
		}

		response[task.Name] = taskInfo
	}

	sendJSON(w, response)
}
