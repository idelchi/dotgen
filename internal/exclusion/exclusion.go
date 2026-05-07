// Package exclusion provides shared exclusion condition handling.
package exclusion

import (
	"fmt"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v4"
)

// Exclude represents one or more exclusion conditions.
type Exclude []bool

// IsExcluded reports whether any exclusion condition is true.
func (e Exclude) IsExcluded() bool {
	for _, condition := range e {
		if condition {
			return true
		}
	}

	return false
}

// UnmarshalYAML parses an exclusion from a bool, string, or list of bool/string values.
func (e *Exclude) UnmarshalYAML(value *yaml.Node) error {
	var conditions []bool

	if err := value.Decode(&conditions); err == nil {
		*e = Exclude(conditions)

		return nil
	}

	var texts []string

	if err := value.Decode(&texts); err == nil {
		conditions, err := parseStrings(texts)
		if err != nil {
			return err
		}

		*e = conditions

		return nil
	}

	var condition bool

	if err := value.Decode(&condition); err == nil {
		*e = Exclude{condition}

		return nil
	}

	var text string

	if err := value.Decode(&text); err == nil {
		condition, err := parseString(text)
		if err != nil {
			return err
		}

		*e = Exclude{condition}

		return nil
	}

	return fmt.Errorf("exclude must be a bool, string, or list of bool/string values")
}

// parseStrings parses string exclusion conditions as booleans.
func parseStrings(values []string) (Exclude, error) {
	conditions := make(Exclude, 0, len(values))
	for _, value := range values {
		condition, err := parseString(value)
		if err != nil {
			return nil, err
		}

		conditions = append(conditions, condition)
	}

	return conditions, nil
}

// parseString parses a single string exclusion condition as a boolean.
func parseString(value string) (bool, error) {
	condition, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, fmt.Errorf("exclude value %q must parse as a bool: %w", value, err)
	}

	return condition, nil
}
