package helpers

import (
	"strings"
	"testing"
)

func TestValidateTaskName(t *testing.T) {
	tests := []struct {
		name        string
		taskName    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid task name with letters",
			taskName:    "ValidName",
			expectError: false,
		},
		{
			name:        "Valid task name with numbers",
			taskName:    "Name123",
			expectError: false,
		},
		{
			name:        "Valid task name with underscore",
			taskName:    "valid_name",
			expectError: false,
		},
		{
			name:        "Valid task name with dash",
			taskName:    "valid-name",
			expectError: false,
		},
		{
			name:        "Valid task name with mixed characters",
			taskName:    "valid-name_123",
			expectError: false,
		},
		{
			name:        "Empty task name",
			taskName:    "",
			expectError: true,
			errorMsg:    "name can not be empty",
		},
		{
			name:        "Task name starting with @",
			taskName:    "@invalid",
			expectError: true,
			errorMsg:    "name can not start with '@'",
		},
		{
			name:        "Task name with invalid characters",
			taskName:    "invalid name",
			expectError: true,
			errorMsg:    "name can only contain Latin letters, numbers, underscores, and dashes",
		},
		{
			name:        "Task name with special characters",
			taskName:    "invalid!name",
			expectError: true,
			errorMsg:    "name can only contain Latin letters, numbers, underscores, and dashes",
		},
		{
			name:        "Task name with non-Latin characters",
			taskName:    "невалидное",
			expectError: true,
			errorMsg:    "name can only contain Latin letters, numbers, underscores, and dashes",
		},
		{
			name:        "Too long task name",
			taskName:    strings.Repeat("a", maxTaskNameLength+1),
			expectError: true,
			errorMsg:    "name can not be longer than",
		},
		{
			name:        "Maximum length task name",
			taskName:    strings.Repeat("a", maxTaskNameLength),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTaskName(tt.taskName)

			// Проверяем, что ошибка соответствует ожиданиям
			if tt.expectError && err == nil {
				t.Errorf("Expected error, but got nil")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}

			// Если ожидается ошибка, проверяем, что её сообщение содержит ожидаемый текст
			if tt.expectError && err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Error message %q does not contain expected text %q", err.Error(), tt.errorMsg)
			}
		})
	}
}

func TestValidateTaskNameEdgeCases(t *testing.T) {
	t.Run("Name with exactly max length", func(t *testing.T) {
		taskName := strings.Repeat("a", maxTaskNameLength)

		err := ValidateTaskName(taskName)
		if err != nil {
			t.Errorf("Expected no error for name with exactly max length, got: %v", err)
		}
	})

	t.Run("Name with exactly max length + 1", func(t *testing.T) {
		taskName := strings.Repeat("a", maxTaskNameLength+1)

		err := ValidateTaskName(taskName)
		if err == nil {
			t.Error("Expected error for name with max length + 1, got nil")
		}
	})

	t.Run("Name with unusual but valid characters", func(t *testing.T) {
		validNames := []string{
			"0123456789",
			"ABCDE",
			"abcde",
			"_-_-_-",
			"abc012",
			"012-_",
			"_012",
			"0",
			"-ABC",
			"-xyz",
			"_0-",
			"-_0z",
		}

		for _, name := range validNames {
			err := ValidateTaskName(name)
			if err != nil {
				t.Errorf("Expected no error for valid name %q, got: %v", name, err)
			}
		}
	})

	t.Run("Name with invalid characters", func(t *testing.T) {
		invalidNames := []string{
			"name with space",
			"name.with.dots",
			"name+with+plus",
			"name/with/slash",
			"name\\with\\backslash",
			"name#with#hash",
			"name@with@at",
		}

		for _, name := range invalidNames {
			err := ValidateTaskName(name)
			if err == nil {
				t.Errorf("Expected error for invalid name %q, got nil", name)
			}
		}
	})
}
