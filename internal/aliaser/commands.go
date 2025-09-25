package aliaser

import (
	"fmt"
	"slices"
	"strings"
)

// Command represents a single command definition, which can be an alias or a function.
type Command struct {
	// Name is the name of the alias or function.
	Name string `yaml:"name"`
	// Cmd is the command or function body.
	Cmd string `yaml:"cmd"`
	// Kind specifies whether it's an "alias" or "function". Defaults to "alias".
	Kind string `yaml:"kind,omitempty"`
	// Shell specifies the shells for which this command is applicable.
	Shell []string `yaml:"shell,omitempty"`
	// OS specifies the operating systems for which this command is applicable.
	OS []string `yaml:"os,omitempty"`
	// Exclude indicates whether to exclude this command entirely.
	Exclude bool `yaml:"exclude,omitempty"`
}

// Export returns a string representation of the command, suitable for shell usage.
func (c *Command) Export() string {
	cmd := strings.TrimSpace(c.Cmd)

	switch c.Kind {
	case "alias":
		return fmt.Sprintf("alias %s='%s'", c.Name, cmd)
	case "function":
		lines := strings.Split(cmd, "\n")
		for i, l := range lines {
			lines[i] = "  " + l
		}

		return fmt.Sprintf("%s() {\n%s\n}", c.Name, strings.Join(lines, "\n"))
	default:
		return fmt.Sprintf("# unknown kind %q for command %q", c.Kind, c.Name)
	}
}

// IsExcluded checks if the command should be excluded based on the provided OS and shell.
func (c *Command) IsExcluded(os, shell string) bool {
	if c.Exclude {
		return true
	}

	if len(c.OS) > 0 && !slices.Contains(c.OS, os) {
		return true
	}

	if len(c.Shell) > 0 && !slices.Contains(c.Shell, shell) {
		return true
	}

	return false
}
