package dotgen

import "github.com/idelchi/dotgen/internal/format"

// Vars represents variables to be set.
type Vars map[string]string

// Export returns a string representation of the variables, suitable for shell usage.
func (v Vars) Export() string {
	return format.Map(v, "%s=%q")
}
