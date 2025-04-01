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
		return fmt.Errorf("name can not be empty")
	}

	// Check if short name starts with @
	if strings.HasPrefix(shortName, "@") {
		return fmt.Errorf("name can not start with '@'")
	}

	// Check if short name contains only allowed characters
	if !shortNameRegex.MatchString(shortName) {
		return fmt.Errorf("name can only contain Latin letters, numbers, underscores, and dashes")
	}

	// Check if short name is too long
	if len(shortName) > maxShortNameLength {
		return fmt.Errorf("name can not be longer than %d characters", maxShortNameLength)
	}

	return nil
}
