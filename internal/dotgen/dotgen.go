package dotgen

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

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

// commandExport holds the rendered output for one command.
type commandExport struct {
	name    string
	command string
	err     error
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
			a.Commands[i].Kind = Alias
		} else if !slices.Contains(Kinds, command.Kind) {
			errs = append(
				errs,
				fmt.Errorf("command %q has invalid kind %q, must be one of %v", command.Name, command.Kind, Kinds),
			)
		}
	}

	return errors.Join(errs...)
}

// Filtered returns a new Dotgen instance with commands filtered based on the provided platforms and shell.
func (a Dotgen) Filtered(platforms []string, shell string) (dotgen Dotgen) {
	for _, c := range a.Commands {
		if c.IsExcluded(platforms, shell) {
			continue
		}

		dotgen.Commands = append(dotgen.Commands, c)
	}

	dotgen.Env = a.Env
	dotgen.Vars = a.Vars

	return dotgen
}

// Export returns a string representation of the Dotgen configuration.
func (a Dotgen) Export(shell, file string, instrument bool, parallel int) (string, error) {
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

	instrumentation := Instrument(file)

	if !instrument {
		instrumentation.Disable()
	}

	buf.WriteString(instrumentation.Header())

	if len(a.Commands) > 0 {
		buf.WriteString("\n# Commands\n")
		buf.WriteString("# ------------------------------------------------\n")

		commands := exportCommands(a.Commands, shell, parallel)

		for _, c := range commands {
			if c.err != nil {
				return "", c.err
			}

			buf.WriteString(instrumentation.Wrap(c.name, c.command))

			buf.WriteString("\n")
		}

		buf.WriteString(instrumentation.Footer())

		buf.WriteString("# ------------------------------------------------\n")
	}

	return strings.TrimSpace(buf.String()), nil
}

// exportCommand renders one command for shell export.
func exportCommand(command Command, shell string) commandExport {
	output, err := command.Export(shell)

	return commandExport{
		name:    command.Name,
		command: output,
		err:     err,
	}
}

// exportCommands renders commands with bounded concurrency while preserving output order.
func exportCommands(commands []Command, shell string, parallel int) []commandExport {
	if parallel < 1 {
		parallel = 1
	}

	exports := make([]commandExport, len(commands))
	if len(commands) == 0 {
		return exports
	}

	if parallel > len(commands) {
		parallel = len(commands)
	}

	if parallel == 1 {
		for i, command := range commands {
			exports[i] = exportCommand(command, shell)
		}

		return exports
	}

	jobs := make(chan int)

	var waitGroup sync.WaitGroup

	for range parallel {
		waitGroup.Go(func() {
			for i := range jobs {
				exports[i] = exportCommand(commands[i], shell)
			}
		})
	}

	for i := range commands {
		jobs <- i
	}

	close(jobs)
	waitGroup.Wait()

	return exports
}
