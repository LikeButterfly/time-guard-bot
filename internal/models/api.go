package models

// Represents the response for task status
type TaskStatusResponse struct {
	Status     string `json:"status"`                // "free", "busy", "locked"
	LockReason string `json:"lock_reason,omitempty"` // Reason for lock if status is "locked"
	TaskName   string `json:"task_name"`             // Name of the task
}

// Represents a task in the list response
type TaskInfo struct {
	ID          string `json:"id"`
	Status      string `json:"status"`                // "free", "busy", "locked"
	LockReason  string `json:"lock_reason,omitempty"` // Only present when status is "locked"
	Description string `json:"description"`
}

// Represents the response for task list
type TaskListResponse map[string]TaskInfo

// Represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
