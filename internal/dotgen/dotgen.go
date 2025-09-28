package dotgen

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"

	"go.yaml.in/yaml/v4"
)

// Dotgen represents the root structure of an dotgen configuration file.
type Dotgen struct {
	// Vars holds the variable definitions.
	Vars Vars `yaml:"vars,omitempty"`
	// Env holds the environment variable definitions.
	Env Env `yaml:"env,omitempty"`
	// Commands holds the command definitions.
	Commands []Command `yaml:"commands,omitempty"`
}

// New parses the provided YAML data into the Dotgen structure.
func New(data []byte) (dotgen Dotgen, err error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))

	dec.KnownFields(true)

	if err := dec.Decode(&dotgen); err != nil {
		return dotgen, fmt.Errorf("parsing alias file: %w", err)
	}

	return dotgen, nil
}

// Validate checks the Dotgen configuration for any issues.
func (a Dotgen) Validate() error {
	errs := []error{}

	for i, command := range a.Commands {
		if command.Kind == "" {
			a.Commands[i].Kind = "alias"
		} else if !slices.Contains(Kinds, command.Kind) {
			errs = append(
				errs,
				fmt.Errorf("command %q has invalid kind %q, must be one of %v", command.Name, command.Kind, Kinds),
			)
		}
	}

	return errors.Join(errs...)
}

// Filtered returns a new Dotgen instance with commands filtered based on the provided OS and shell.
func (a Dotgen) Filtered(os, shell string) (dotgen Dotgen) {
	for _, c := range a.Commands {
		if c.IsExcluded(os, shell) {
			continue
		}

		dotgen.Commands = append(dotgen.Commands, c)
	}

	dotgen.Env = a.Env
	dotgen.Vars = a.Vars

	return dotgen
}

// Export returns a string representation of the Dotgen configuration.
func (a Dotgen) Export(shell string) (string, error) {
	var buf bytes.Buffer

	if len(a.Env) > 0 {
		buf.WriteString("\n# Environment variables\n")
		buf.WriteString("# ------------------------------------------------\n")
		buf.WriteString(a.Env.Export())
		buf.WriteString("\n")
		buf.WriteString("# ------------------------------------------------\n")
	}

	if len(a.Vars) > 0 {
		buf.WriteString("\n# Variables\n")
		buf.WriteString("# ------------------------------------------------\n")
		buf.WriteString(a.Vars.Export())
		buf.WriteString("\n")
		buf.WriteString("# ------------------------------------------------\n")
	}

	if len(a.Commands) > 0 {
		buf.WriteString("\n# Commands\n")
		buf.WriteString("# ------------------------------------------------\n")

		for _, c := range a.Commands {
			command, err := c.Export(shell)
			if err != nil {
				return "", fmt.Errorf("exporting command %q: %w", c.Name, err)
			}

			buf.WriteString(command)
			buf.WriteString("\n")
		}

		buf.WriteString("# ------------------------------------------------\n")
	}

	return strings.TrimSpace(buf.String()), nil
}
