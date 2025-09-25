// Package cli implements the command-line interface for aliaser.
//
//nolint:forbidigo // This package prints out to the console.
package cli

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/pflag"

	"github.com/idelchi/aliaser/internal/aliaser"
	"github.com/idelchi/aliaser/internal/variables"
	"github.com/idelchi/aliaser/pkg/template"
)

// CLI represents the command-line interface.
type CLI struct {
	version string
}

// New creates a new CLI instance with the given version.
func New(version string) CLI {
	return CLI{version: version}
}

func help() {
	fmt.Println(heredoc.Doc(`
		aliaser is a tool to manage and execute named shell commands with Go template substitution.

		Usage:

			aliaser [flags][key=value ...]

		Flags:
	`))
	pflag.PrintDefaults()
}

// Execute runs the CLI with the provided arguments.
func (c CLI) Execute() error {
	defaultConfig := os.ExpandEnv(os.Getenv("ALIASER_CONFIG"))
	defaultShell := os.Getenv("SHELL")

	if defaultConfig == "" {
		if home, err := os.UserHomeDir(); err == nil {
			defaultConfig = filepath.ToSlash(filepath.Join(home, ".config", "aliaser", "aliaser.yaml"))
		}
	}

	defaultConfig = filepath.ToSlash(defaultConfig)

	if defaultShell != "" {
		defaultShell = filepath.Base(defaultShell)
	} else {
		defaultShell = "zsh"
	}

	var (
		config  = pflag.StringP("config", "c", defaultConfig, "Path to the aliaser configuration file")
		shell   = pflag.StringP("shell", "s", defaultShell, "The active shell")
		values  = pflag.StringSliceP("values", "v", []string{}, "Additional YAML value files")
		show    = pflag.BoolP("show", "S", false, "Show the collected variables and exit")
		version = pflag.BoolP("version", "V", false, "Show the version and exit")
	)

	pflag.CommandLine.SortFlags = false

	pflag.Usage = help
	pflag.Parse()

	if *version {
		fmt.Println(c.version)

		return nil
	}

	vars := variables.Defaults(*shell)

	if values, err := variables.Values(*values).Variables(); err == nil {
		maps.Copy(vars, values)
	} else {
		return err //nolint:wrapcheck // Error is already descriptive enough
	}

	if values, err := variables.Args(pflag.Args()).ToKeyValues(); err == nil {
		maps.Copy(vars, values)
	} else {
		return fmt.Errorf("parsing args: %w", err)
	}

	if *show {
		fmt.Println(vars.Export())

		return nil
	}

	if *config == "" {
		return errors.New("no config file provided, use --config or set ALIASER_CONFIG")
	}

	data, err := os.ReadFile(*config)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	rendered, err := template.Apply(string(data), vars)
	if err != nil {
		return err //nolint:wrapcheck // Error is already descriptive enough
	}

	aliaser, err := aliaser.New([]byte(rendered))
	if err != nil {
		return err //nolint:wrapcheck // Error is already descriptive enough
	}

	if err := aliaser.Validate(); err != nil {
		return err //nolint:wrapcheck // Error is already descriptive enough
	}

	os, ok := vars["OS"].(string)
	if !ok {
		return fmt.Errorf("expected string for OS, got %T", vars["OS"])
	}

	aliaser = aliaser.Filtered(os, *shell)

	fmt.Println(aliaser.Export())

	return nil
}
