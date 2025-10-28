// Package cli implements the command-line interface for dotgen.
package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
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
//
//nolint:gocognit // Lengthy command setup.
func (c CLI) Execute() error {
	var options Options

	var completion string

	shell := filepath.Base(os.Getenv("SHELL"))
	if shell == "." {
		shell = ""
	}

	root := &cobra.Command{
		Use:   "dotgen [flags] [patterns ...]",
		Short: "Manage and execute named shell commands with Go template substitution",
		Long: heredoc.Docf(`
			dotgen is a tool to manage and execute named shell commands with Go template substitution.

			Positional Arguments:
			  patterns               Paths or patterns to dotgen configuration files. Defaults to %q if not specified.
		`, DefaultPath),
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       c.version,
		RunE: func(cmd *cobra.Command, args []string) error {
			if completion != "" {
				return completions(cmd, completion)
			}

			if len(args) == 0 {
				options.Input = []string{DefaultPath}
			} else {
				options.Input = args
			}

			for idx, pattern := range options.Input {
				pattern = filepath.ToSlash(pattern)

				switch {
				case pattern == ".":
					pattern = DefaultPath
				case strings.HasSuffix(pattern, "/"):
					pattern = filepath.Join(pattern, DefaultPath)
				default:
					info, err := os.Stat(pattern)
					if err == nil && info.IsDir() {
						pattern = filepath.Join(pattern, DefaultPath)
					}
				}

				options.Input[idx] = filepath.ToSlash(pattern)
			}

			if options.Shell == "" {
				return errors.New("no shell specified, provide using --shell or SHELL environment variable")
			}

			if options.Debug || options.Instrument {
				options.Verbose = true
			}

			logger := Logger{Verbose: options.Verbose}

			return logic(options, logger)
		},
	}

	root.Flags().StringVar(&options.Shell, "shell", shell, "The active shell")
	root.Flags().StringSliceVarP(&options.Values, "values", "f", []string{}, "Additional YAML value files")
	root.Flags().
		StringSliceVar(&options.Set, "set", []string{}, "Set or override variables (key=value), strings only")
	root.Flags().BoolVar(&options.Verbose, "verbose", false, "Show verbose output")
	root.Flags().BoolVar(&options.Debug, "debug", false, "Show debug output")
	root.Flags().BoolVarP(&options.Instrument, "instrument", "I", false, "Enable instrumentation for profiling")
	root.Flags().
		BoolVar(&options.Hash, "hash", false, "Compute a hash of all files that would be included and print it out")
	root.Flags().
		BoolVar(&options.Dry, "dry", false, "Show which files would be included, but do not execute commands")
	root.Flags().
		StringVar(&completion, "shell-completion", "",
			"Generate shell completion script for specified shell (bash|zsh|fish|powershell)")

	_ = root.Flags().MarkHidden("shell-completion")

	root.Flags().SortFlags = false

	return root.Execute() //nolint:wrapcheck 	// Error does not need additional wrapping.
}
