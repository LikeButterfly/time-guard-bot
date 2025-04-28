package helpers

import (
	"strings"
	"testing"
)

func TestValidateShortName(t *testing.T) {
	tests := []struct {
		name        string
		shortName   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid short name with letters",
			shortName:   "ValidName",
			expectError: false,
		},
		{
			name:        "Valid short name with numbers",
			shortName:   "Name123",
			expectError: false,
		},
		{
			name:        "Valid short name with underscore",
			shortName:   "valid_name",
			expectError: false,
		},
		{
			name:        "Valid short name with dash",
			shortName:   "valid-name",
			expectError: false,
		},
		{
			name:        "Valid short name with mixed characters",
			shortName:   "valid-name_123",
			expectError: false,
		},
		{
			name:        "Empty short name",
			shortName:   "",
			expectError: true,
			errorMsg:    "name can not be empty",
		},
		{
			name:        "Short name starting with @",
			shortName:   "@invalid",
			expectError: true,
			errorMsg:    "name can not start with '@'",
		},
		{
			name:        "Short name with invalid characters",
			shortName:   "invalid name",
			expectError: true,
			errorMsg:    "name can only contain Latin letters, numbers, underscores, and dashes",
		},
		{
			name:        "Short name with special characters",
			shortName:   "invalid!name",
			expectError: true,
			errorMsg:    "name can only contain Latin letters, numbers, underscores, and dashes",
		},
		{
			name:        "Short name with non-Latin characters",
			shortName:   "невалидное",
			expectError: true,
			errorMsg:    "name can only contain Latin letters, numbers, underscores, and dashes",
		},
		{
			name:        "Too long short name",
			shortName:   strings.Repeat("a", maxShortNameLength+1),
			expectError: true,
			errorMsg:    "name can not be longer than",
		},
		{
			name:        "Maximum length short name",
			shortName:   strings.Repeat("a", maxShortNameLength),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateShortName(tt.shortName)

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

func TestValidateShortNameEdgeCases(t *testing.T) {
	t.Run("Name with exactly max length", func(t *testing.T) {
		shortName := strings.Repeat("a", maxShortNameLength)

		err := ValidateShortName(shortName)
		if err != nil {
			t.Errorf("Expected no error for name with exactly max length, got: %v", err)
		}
	})

	t.Run("Name with exactly max length + 1", func(t *testing.T) {
		shortName := strings.Repeat("a", maxShortNameLength+1)

		err := ValidateShortName(shortName)
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
			err := ValidateShortName(name)
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
			err := ValidateShortName(name)
			if err == nil {
				t.Errorf("Expected error for invalid name %q, got nil", name)
			}
		}
	})
}
