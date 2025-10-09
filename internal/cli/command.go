// Package cli implements the command-line interface for dotgen.
//
//nolint:forbidigo // This package prints out to the console.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// DefaultPath is the default glob pattern for dotgen configuration files.
const DefaultPath = "**/*.dotgen"

func help() {
	fmt.Println(heredoc.Docf(`
		dotgen is a tool to manage and execute named shell commands with Go template substitution.

		Usage:

			dotgen [flags] [patterns ...]

		Positional Arguments:
		  patterns               Paths or patterns to dotgen configuration files. Defaults to %q if not specified.

		Flags:
	`, DefaultPath))
	pflag.PrintDefaults()
}

// Options represents the CLI options.
type Options struct {
	// Input represents input file paths or patterns.
	Input []string
	// Shell represents the active shell.
	Shell string
	// Values represents additional YAML value files.
	Values []string
	// Set represents variables to set or override (key=value).
	Set []string
	// Verbose represents whether verbose output is enabled.
	Verbose bool
	// Debug represents whether debug output is enabled.
	Debug bool
	// Instrument represents whether instrumentation for profiling is enabled.
	Instrument bool
	// Hash represents whether to compute and print a hash of all included files.
	Hash bool
	// Dry represents whether to show which files would be included, but not execute commands.
	Dry bool
	// Version represents whether to show the version and exit.
	Version bool
}

// Execute runs the CLI with the provided arguments.
func (c CLI) Execute() error {
	var options Options

	shell := filepath.Base(os.Getenv("SHELL"))
	if shell == "." {
		shell = ""
	}

	pflag.StringVar(&options.Shell, "shell", shell, "The active shell")
	pflag.StringSliceVarP(&options.Values, "values", "f", []string{}, "Additional YAML value files")
	pflag.StringSliceVar(&options.Set, "set", []string{}, "Set or override variables (key=value), strings only")
	pflag.BoolVar(&options.Verbose, "verbose", false, "Show verbose output")
	pflag.BoolVar(&options.Debug, "debug", false, "Show debug output")
	pflag.BoolVarP(&options.Instrument, "instrument", "I", false, "Enable instrumentation for profiling")
	pflag.BoolVar(&options.Hash, "hash", false, "Compute a hash of all files that would be included and print it out")
	pflag.BoolVar(&options.Dry, "dry", false, "Show which files would be included, but do not execute commands")
	pflag.BoolVarP(&options.Version, "version", "v", false, "Show the version and exit")

	pflag.CommandLine.SortFlags = false

	pflag.Usage = help
	pflag.Parse()

	if pflag.NArg() == 0 {
		options.Input = []string{DefaultPath}
	} else {
		options.Input = pflag.Args()
	}

	for idx, pattern := range options.Input {
		pattern = filepath.ToSlash(pattern)

		if pattern == "." {
			pattern = DefaultPath
		} else if strings.HasSuffix(pattern, "/") {
			pattern = filepath.Join(pattern, DefaultPath)
		}

		options.Input[idx] = filepath.ToSlash(pattern)
	}

	if options.Version {
		fmt.Println(c.version)

		return nil
	}

	if options.Shell == "" {
		return fmt.Errorf("no shell specified, provide using --shell/-s or set SHELL environment variable")
	}

	if options.Debug || options.Instrument {
		options.Verbose = true
	}

	logger := Logger{Verbose: options.Verbose}

	return logic(options, logger)
}
