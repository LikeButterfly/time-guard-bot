// Copyright 2025 LikeButterfly
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package helpers

import (
	"fmt"
	"strings"
	"testing"
)

func TestGenerateTaskID(t *testing.T) {
	// Тестируем разные длины ID
	lengths := []int{5, 6, 7, 8, 9, 10}

	for _, length := range lengths {
		testName := fmt.Sprintf("Length_%d", length)
		t.Run(testName, func(t *testing.T) {
			taskID, err := GenerateTaskID(length)

			// Проверяем, что нет ошибки
			if err != nil {
				t.Fatalf("GenerateTaskID(%d) returned error: %v", length, err)
			}

			// Проверяем длину сгенерированного ID
			if len(taskID) != length {
				t.Errorf("Expected ID of length %d, got %d: %s", length, len(taskID), taskID)
			}

			// Проверяем, что ID содержит только допустимые символы
			for _, char := range taskID {
				if !strings.ContainsRune(TaskIDChars, char) {
					t.Errorf("ID contains invalid character: %c", char)
					break
				}
			}
		})
	}
}

func TestGenerateTaskIDRandomness(t *testing.T) {
	// Проверяем, что генерация двух ID с одинаковой длиной дает разные результаты
	length := TaskIDLength

	// Генерируем множество ID, чтобы проверить случайность
	idMap := make(map[string]bool)
	iterations := 100

	for i := range iterations {
		id, err := GenerateTaskID(length)
		if err != nil {
			t.Fatalf("GenerateTaskID(%d) returned error on iteration %d: %v", length, i, err)
		}

		// Проверяем, что мы не сгенерировали такой же ID ранее
		if idMap[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}

		idMap[id] = true
	}

	// Проверяем, что у нас правильное количество уникальных ID
	if len(idMap) != iterations {
		t.Errorf("Expected %d unique IDs, got %d", iterations, len(idMap))
	}
}

func TestGenerateTaskIDWithConstants(t *testing.T) {
	t.Run("Using TaskIDLength constant", func(t *testing.T) {
		taskID, err := GenerateTaskID(TaskIDLength)
		if err != nil {
			t.Fatalf("GenerateTaskID(%d) returned error: %v", TaskIDLength, err)
		}

		if len(taskID) != TaskIDLength {
			t.Errorf("Expected ID length %d, got %d: %s", TaskIDLength, len(taskID), taskID)
		}
	})
}
