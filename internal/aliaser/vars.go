package aliaser

import (
	"fmt"
	"strings"
)

// Vars represents variables to be set.
type Vars map[string]string

// Export returns a string representation of the variables, suitable for shell usage.
func (v Vars) Export() string {
	out := make([]string, 0, len(v))
	for k, val := range v {
		out = append(out, fmt.Sprintf("%s=%s", k, val))
	}

	return strings.Join(out, "\n")
}
