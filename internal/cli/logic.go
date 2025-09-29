package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/idelchi/dotgen/internal/dotgen"
	"github.com/idelchi/dotgen/internal/format"
	"github.com/idelchi/dotgen/internal/split"
	"github.com/idelchi/dotgen/internal/variables"
	"github.com/idelchi/dotgen/pkg/template"
)

// logic contains the main logic of the CLI.
//
// It processes the provided options, loads and merges variables, reads and
// renders dotgen configuration files, validates them, filters commands based
// on the current OS and shell, and exports the final configuration to the
// console.
//
//nolint:gocognit,funlen,forbidigo // TODO(Idelchi): Refactor
func logic(options Options, logger Logger) error {
	if options.Debug {
		fmt.Println("default variables:")
		fmt.Println("*******************")
		fmt.Println(format.Map(variables.Defaults(options.Shell, ""), "%s=%q"))
		fmt.Println("*******************")
		fmt.Println()
	}

	if len(options.Input) == 0 {
		return errors.New("no input file provided, specify using --input/-i")
	}

	files := []string{}

	for _, file := range options.Input {
		file = format.Path(file)

		base, pattern := doublestar.SplitPattern(file)
		fsys := os.DirFS(base)

		logger.Printlnf("loading config %q", file)

		matches, err := doublestar.Glob(fsys, pattern)
		if err != nil {
			return fmt.Errorf("invalid config pattern %q: %w", file, err)
		}

		logger.Printlnf(" - found %d file(s) for pattern %q", len(matches), file)

		for _, file := range matches {
			file = filepath.ToSlash(filepath.Join(base, file))
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		return fmt.Errorf("no files matched the provided patterns: %v", options.Input)
	}

	logger.Printlnf("processing %d file(s)", len(files))

	logger.Printlnf(" - processing:")

	for _, file := range files {
		logger.Printlnf("  - %q", file)

		vars, err := mergeVars(options, nil, file)
		if err != nil {
			return err
		}

		operatingSystem := getOSExtensionFromFileName(file)

		if operatingSystem != "" && operatingSystem != vars["OS"] {
			logger.Printlnf(
				"    - skipping due to file suffix OS exclusion: file is for %q, current OS is %q",
				operatingSystem,
				vars["OS"],
			)

			continue
		}

		data, err := os.ReadFile(filepath.Clean(file))
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		docs := split.YAML(data)

		var doc []byte

		const maxDocs = 2

		switch len(docs) {
		case 0:
			continue
		case 1:
			doc = docs[0]
		case maxDocs:
			rendered, err := template.Apply(string(docs[0]), vars)
			if err != nil {
				return err //nolint:wrapcheck // Error is already descriptive enough
			}

			if options.Debug {
				fmt.Println("header rendered as:")
				fmt.Println("*******************")
				fmt.Println(rendered)
				fmt.Println("*******************")
				fmt.Println()
			}

			header, err := variables.NewHeader([]byte(rendered))
			if err != nil {
				return fmt.Errorf("parsing header in %q: %w", file, err)
			}

			if header.Exclude {
				logger.Printlnf("    - skipping due to header exclusion")

				continue
			}

			vars, err = mergeVars(options, header.Values, file)
			if err != nil {
				return err
			}

			if options.Debug {
				fmt.Println("merged variables:")
				fmt.Println("*******************")
				fmt.Println(format.Map(vars, "%s=%q"))
				fmt.Println("*******************")
			}

			doc = docs[1]
		default:
			return fmt.Errorf("expected at most 2 documents in %q, got %d", file, len(docs))
		}

		rendered, err := template.Apply(string(doc), vars)
		if err != nil {
			return err //nolint:wrapcheck // Error is already descriptive enough
		}

		if options.Debug {
			fmt.Println("body rendered as:")
			fmt.Println("*******************")
			fmt.Println(rendered)
			fmt.Println("*******************")

			continue
		}

		dotgen, err := dotgen.New([]byte(rendered))
		if err != nil {
			return err //nolint:wrapcheck // Error is already descriptive enough
		}

		if err := dotgen.Validate(); err != nil {
			return err //nolint:wrapcheck // Error is already descriptive enough
		}

		os, ok := vars["OS"].(string)
		if !ok {
			return fmt.Errorf("expected string for OS, got %T", vars["OS"])
		}

		dotgen = dotgen.Filtered(os, options.Shell)

		export, err := dotgen.Export(options.Shell, file, options.Instrument)
		if err != nil {
			return fmt.Errorf("exporting dotgen config: %w", err)
		}

		if options.Verbose {
			// Build the line first
			line := fmt.Sprintf("# Generated from %q", file)

			// Repeat stars to the length of the line
			stars := strings.Repeat("*", len(line))

			fmt.Println("# " + stars)
			fmt.Println(line)
			fmt.Println("# " + stars)

			fmt.Println("# Template variables:")
			fmt.Println("# " + stars)
			fmt.Println(format.Map(vars, "# %s=%q"))
			fmt.Println("# " + stars)
			fmt.Println()
		}

		fmt.Println(export)
		fmt.Println()
	}

	return nil
}
