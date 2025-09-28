package cli

import (
	"fmt"
	"os"
)

// Logger represents a simple logger with verbosity control.
type Logger struct {
	// Verbose indicates whether verbose logging is enabled.
	Verbose bool
}

// Printlnf prints a formatted message to stderr if Verbose is true.
func (l Logger) Printlnf(format string, args ...any) {
	if l.Verbose {
		fmt.Fprintf(os.Stderr, "[dotgen]: "+format+"\n", args...)
	}
}
