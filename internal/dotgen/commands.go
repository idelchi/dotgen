package dotgen

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/idelchi/dotgen/internal/format"
	"github.com/idelchi/dotgen/pkg/exec"
)

// Command represents a single command definition, which can be an alias or a function.
type Command struct {
	// Name is the name of the command.
	Name string `yaml:"name"`
	// Doc is the documentation string for the command.
	Doc string `yaml:"doc,omitempty"`
	// Cmd is the command to execute.
	Cmd string `yaml:"cmd"`
	// Kind is the type of command: "alias", "function", "raw", or "run".
	Kind string `yaml:"kind,omitempty"`
	// ExportTo is the path to export the command output.
	ExportTo string `yaml:"export_to,omitempty"`
	// Shell specifies the shells for which this command is applicable.
	Shell []string `yaml:"shell,omitempty"`
	// OS specifies the operating systems for which this command is applicable.
	OS []string `yaml:"os,omitempty"`
	// Exclude specifies whether to exclude this command from the output.
	Exclude bool `yaml:"exclude,omitempty"`
}

// Export returns a string representation of the command, suitable for shell usage.
//
//nolint:gocognit,funlen // TODO(Idelchi): Refactor
func (c *Command) Export(shell string) (string, error) {
	name := strings.TrimSpace(c.Name)
	cmd := strings.TrimSpace(c.Cmd)
	doc := strings.ReplaceAll(strings.TrimSpace(c.Doc), "\n", "\n#  ")

	switch c.Kind {
	case "alias":
		var builder strings.Builder
		fmt.Fprintf(&builder, "# name: %s\n", name)

		if c.Doc != "" {
			fmt.Fprint(&builder, "# doc:\n")
			fmt.Fprintf(&builder, "#  %s\n", doc)
		}

		formatted, err := format.Shell(cmd, true)
		if err == nil {
			cmd = formatted
		}

		fmt.Fprintf(&builder, "alias %s='%s'", name, strings.TrimRight(cmd, "\n"))
		fmt.Fprint(&builder, "\n")

		return builder.String(), nil
	case "function":
		function := fmt.Sprintf("%s() {\n%s\n}\n", name, cmd)

		var builder strings.Builder

		fmt.Fprintf(&builder, "# name: %s\n", name)

		if c.Doc != "" {
			fmt.Fprint(&builder, "# doc:\n")
			fmt.Fprintf(&builder, "#  %s\n", doc)
		}

		fmt.Fprint(&builder, function)

		function = builder.String()

		formatted, err := format.Shell(function, false)
		if err == nil {
			function = formatted
		}

		return function, nil
	case "raw":
		raw := c.Cmd

		var builder strings.Builder

		fmt.Fprintf(&builder, "# name: %s\n", name)

		if c.Doc != "" {
			fmt.Fprint(&builder, "# doc:\n")
			fmt.Fprintf(&builder, "#  %s\n", doc)
		}

		fmt.Fprint(&builder, raw)

		raw = builder.String()

		formatted, err := format.Shell(raw, false)
		if err == nil {
			raw = formatted
		}

		return raw, nil

	case "run":
		result := exec.Run(
			shell,
			cmd,
		)

		if result.Err != nil {
			return "", fmt.Errorf("executing command %q: %w: %v", name, result.Err, result.Stderr)
		}

		var builder strings.Builder

		fmt.Fprintf(&builder, "# name: %s\n", name)

		if c.Doc != "" {
			fmt.Fprintf(&builder, "# doc:\n")
			fmt.Fprintf(&builder, "#  %s\n", c.Doc)
		}

		cmd = strings.ReplaceAll(strings.TrimSpace(cmd), "\n", "\n#  ")

		fmt.Fprint(&builder, "# original:\n")
		fmt.Fprintf(&builder, "#  %s\n", cmd)

		if c.ExportTo != "" {
			exportTo := os.ExpandEnv(c.ExportTo)
			fmt.Fprintf(&builder, "# output exported to %q\n", exportTo)
			fmt.Fprintf(&builder, ". %q\n", exportTo)

			if err := os.MkdirAll(filepath.Dir(exportTo), 0o700); err != nil {
				return "", fmt.Errorf("creating directories for %q: %w", exportTo, err)
			}

			if err := os.WriteFile(exportTo, []byte(result.Stdout), 0o600); err != nil {
				return "", fmt.Errorf("writing output to %q: %w", exportTo, err)
			}
		} else {
			builder.WriteString(result.Stdout)
		}

		return builder.String(), nil

	default:
		return "", fmt.Errorf("# unknown kind %q for command %q", c.Kind, c.Name)
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
