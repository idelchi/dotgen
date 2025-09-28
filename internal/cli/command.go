// Package cli implements the command-line interface for dotgen.
//
//nolint:forbidigo // This package prints out to the console.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/pflag"
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
		dotgen is a tool to manage and execute named shell commands with Go template substitution.

		Usage:

			dotgen [flags] [key=value ...]

		Flags:
	`))
	pflag.PrintDefaults()
}

// Options represents the CLI options.
type Options struct {
	// Input represents the input configuration files.
	// Processed as doublestar patterns.
	Input []string
	// Shell represents the active shell.
	Shell string
	// Values represents additional YAML value files.
	Values []string
	// Verbose represents whether verbose output is enabled.
	Verbose bool
	// Debug represents whether debug output is enabled.
	Debug bool
	// Version represents whether to show the version and exit.
	Version bool
}

// Execute runs the CLI with the provided arguments.
func (c CLI) Execute() error {
	defaultShell := os.Getenv("SHELL")

	if defaultShell != "" {
		defaultShell = filepath.Base(defaultShell)
	} else {
		defaultShell = "zsh"
	}

	var options Options

	pflag.StringSliceVarP(&options.Input, "input", "i", []string{}, "Paths or patterns to dotgen configuration files")
	pflag.StringVarP(&options.Shell, "shell", "s", defaultShell, "The active shell")
	pflag.StringSliceVarP(&options.Values, "values", "V", []string{}, "Additional YAML value files")
	pflag.BoolVar(&options.Verbose, "verbose", false, "Show verbose output")
	pflag.BoolVarP(&options.Debug, "debug", "d", false, "Show debug output")
	pflag.BoolVarP(&options.Version, "version", "v", false, "Show the version and exit")

	pflag.CommandLine.SortFlags = false

	pflag.Usage = help
	pflag.Parse()

	if options.Version {
		fmt.Println(c.version)

		return nil
	}

	if options.Debug {
		options.Verbose = true
	}

	logger := Logger{Verbose: options.Verbose}

	return logic(options, logger)
}
