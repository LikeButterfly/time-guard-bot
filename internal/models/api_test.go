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

import (
	"encoding/json"
	"testing"
)

func TestAPIStructures(t *testing.T) {
	t.Run("TaskStatusResponse", func(t *testing.T) {
		resp := TaskStatusResponse{
			Status:     "locked",
			LockReason: "Maintenance",
			TaskName:   "Test_Task",
		}

		// Проверяем маршализацию в JSON
		jsonBytes, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("Failed to marshal TaskStatusResponse: %v", err)
		}

		// Проверяем демаршализацию
		var unmarshaled TaskStatusResponse

		err = json.Unmarshal(jsonBytes, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal TaskStatusResponse: %v", err)
		}

		// Проверяем поля
		if resp.Status != unmarshaled.Status {
			t.Errorf("Status mismatch: expected %s, got %s", resp.Status, unmarshaled.Status)
		}

		if resp.LockReason != unmarshaled.LockReason {
			t.Errorf("LockReason mismatch: expected %s, got %s", resp.LockReason, unmarshaled.LockReason)
		}

		if resp.TaskName != unmarshaled.TaskName {
			t.Errorf("TaskName mismatch: expected %s, got %s", resp.TaskName, unmarshaled.TaskName)
		}
	})

	t.Run("TaskInfo", func(t *testing.T) {
		info := TaskInfo{
			ID:          "12345",
			Status:      "locked",
			LockReason:  "Maintenance",
			Description: "Test_task",
		}

		// Проверяем маршализацию в JSON
		jsonBytes, err := json.Marshal(info)
		if err != nil {
			t.Fatalf("Failed to marshal TaskInfo: %v", err)
		}

		// Проверяем демаршализацию
		var unmarshaled TaskInfo

		err = json.Unmarshal(jsonBytes, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal TaskInfo: %v", err)
		}

		// Проверяем поля
		if info.ID != unmarshaled.ID {
			t.Errorf("ID mismatch: expected %s, got %s", info.ID, unmarshaled.ID)
		}

		if info.Status != unmarshaled.Status {
			t.Errorf("Status mismatch: expected %s, got %s", info.Status, unmarshaled.Status)
		}

		if info.LockReason != unmarshaled.LockReason {
			t.Errorf("LockReason mismatch: expected %s, got %s", info.LockReason, unmarshaled.LockReason)
		}

		if info.Description != unmarshaled.Description {
			t.Errorf("Description mismatch: expected %s, got %s", info.Description, unmarshaled.Description)
		}
	})

	t.Run("TaskListResponse", func(t *testing.T) {
		response := TaskListResponse{
			"Task 1": TaskInfo{
				ID:          "task1",
				Status:      "free",
				Description: "First task",
			},
			"Task 2": TaskInfo{
				ID:          "task2",
				Status:      "locked",
				LockReason:  "Maintenance",
				Description: "Second task",
			},
		}

		// Проверяем маршализацию в JSON
		jsonBytes, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("Failed to marshal TaskListResponse: %v", err)
		}

		// Проверяем демаршализацию
		var unmarshaled TaskListResponse

		err = json.Unmarshal(jsonBytes, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal TaskListResponse: %v", err)
		}

		// Проверяем количество задач
		if len(unmarshaled) != len(response) {
			t.Errorf("Number of tasks mismatch: expected %d, got %d", len(response), len(unmarshaled))
		}

		// Проверяем содержимое каждой задачи
		for name, info := range response {
			unmarshaledInfo, exists := unmarshaled[name]
			if !exists {
				t.Errorf("Task '%s' not found in unmarshaled response", name)
				continue
			}

			if info.ID != unmarshaledInfo.ID {
				t.Errorf("ID mismatch for task '%s': expected %s, got %s", name, info.ID, unmarshaledInfo.ID)
			}

			if info.Status != unmarshaledInfo.Status {
				t.Errorf("Status mismatch for task '%s': expected %s, got %s", name, info.Status, unmarshaledInfo.Status)
			}

			if info.LockReason != unmarshaledInfo.LockReason {
				t.Errorf("LockReason mismatch for task '%s': expected %s, got %s", name, info.LockReason, unmarshaledInfo.LockReason)
			}
		}
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		errResp := ErrorResponse{
			Error: "Something went wrong",
		}

		// Проверяем маршализацию в JSON
		jsonBytes, err := json.Marshal(errResp)
		if err != nil {
			t.Fatalf("Failed to marshal ErrorResponse: %v", err)
		}

		// Проверяем демаршализацию
		var unmarshaled ErrorResponse

		err = json.Unmarshal(jsonBytes, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal ErrorResponse: %v", err)
		}

		// Проверяем поля
		if errResp.Error != unmarshaled.Error {
			t.Errorf("Error message mismatch: expected %s, got %s", errResp.Error, unmarshaled.Error)
		}
	})
}
