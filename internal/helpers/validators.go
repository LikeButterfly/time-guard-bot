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
package helpers

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// Maximum allowed characters for task name
	maxTaskNameLength = 16 // TODO мб вынести?
)

var (
	// Regexp for validating name (letters, numbers, underscore, dash)
	taskNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// Helper function to validate task name
func ValidateTaskName(name string) error {
	// Check if name is empty
	if name == "" {
		return fmt.Errorf("name can not be empty")
	}

	// Check if name starts with @
	if strings.HasPrefix(name, "@") {
		return fmt.Errorf("name can not start with '@'")
	}

	// Check if name contains only allowed characters
	if !taskNameRegex.MatchString(name) {
		return fmt.Errorf("name can only contain Latin letters, numbers, underscores, and dashes")
	}

	// Check if name is too long
	if len(name) > maxTaskNameLength {
		return fmt.Errorf("name can not be longer than %d characters", maxTaskNameLength)
	}

	return nil
}
