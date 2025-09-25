package aliaser

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"

	"go.yaml.in/yaml/v4"
)

// Aliaser represents the root structure of an aliaser configuration file.
type Aliaser struct {
	// Vars holds the variable definitions.
	Vars Vars `yaml:"vars,omitempty"`
	// Env holds the environment variable definitions.
	Env Env `yaml:"env,omitempty"`
	// Commands holds the command definitions.
	Commands []Command `yaml:"commands,omitempty"`
}

// New parses the provided YAML data into the Aliaser structure.
func New(data []byte) (aliaser Aliaser, err error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))

	dec.KnownFields(true)

	if err := dec.Decode(&aliaser); err != nil {
		return aliaser, fmt.Errorf("parsing alias file: %w", err)
	}

	return aliaser, nil
}

// Validate checks the Aliaser configuration for any issues.
func (a Aliaser) Validate() error {
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

// Filtered returns a new Aliaser instance with commands filtered based on the provided OS and shell.
func (a Aliaser) Filtered(os, shell string) (aliaser Aliaser) {
	for _, c := range a.Commands {
		if c.IsExcluded(os, shell) {
			continue
		}

		aliaser.Commands = append(aliaser.Commands, c)
	}

	aliaser.Env = a.Env
	aliaser.Vars = a.Vars

	return aliaser
}

// Export returns a string representation of the Aliaser configuration.
func (a Aliaser) Export() string {
	var buf bytes.Buffer

	if len(a.Env) > 0 {
		buf.WriteString("\n# Environment variables\n")
		buf.WriteString(a.Env.Export())
		buf.WriteString("\n")
	}

	if len(a.Vars) > 0 {
		buf.WriteString("\n# Variables\n")
		buf.WriteString(a.Vars.Export())
		buf.WriteString("\n")
	}

	if len(a.Commands) > 0 {
		buf.WriteString("\n# Commands\n")

		for _, c := range a.Commands {
			buf.WriteString(c.Export())
			buf.WriteString("\n")
		}
	}

	return strings.TrimSpace(buf.String())
}
