// Copyright 2025 LikeButterfly
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
