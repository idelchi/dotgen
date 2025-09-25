package aliaser

import (
	"fmt"
	"strings"
)

// Env represents environment variables to be set.
type Env map[string]string

// Export returns a string representation of the environment variables, suitable for shell usage.
func (e Env) Export() string {
	out := make([]string, 0, len(e))
	for k, val := range e {
		out = append(out, fmt.Sprintf("export %s=%s", k, val))
	}

	return strings.Join(out, "\n")
}
