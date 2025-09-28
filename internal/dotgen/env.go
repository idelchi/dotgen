package dotgen

import "github.com/idelchi/dotgen/internal/format"

// Env represents environment variables to be set.
type Env map[string]string

// Export returns a string representation of the environment variables, suitable for shell usage.
func (e Env) Export() string {
	return format.Map(e, "export %s=%q")
}
