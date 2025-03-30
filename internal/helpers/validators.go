package helpers

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// Maximum allowed characters for task short name
	maxShortNameLength = 16 // TODO мб вынести?
)

var (
	// Regexp for validating short name (letters, numbers, underscore, dash)
	shortNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// Helper function to validate task short name
func ValidateShortName(shortName string) error {
	// Check if short name is empty
	if shortName == "" {
		return fmt.Errorf("краткое название не может быть пустым")
	}

	// Check if short name starts with @
	if strings.HasPrefix(shortName, "@") {
		return fmt.Errorf("краткое название не может начинаться с '@'")
	}

	// Check if short name is too long
	if len(shortName) > maxShortNameLength {
		return fmt.Errorf("краткое название не может быть длиннее %d символов", maxShortNameLength)
	}

	// Check if short name contains only allowed characters
	if !shortNameRegex.MatchString(shortName) {
		return fmt.Errorf("краткое название может содержать только буквы, цифры, подчеркивания и дефисы")
	}

	return nil
}
