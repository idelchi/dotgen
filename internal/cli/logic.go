package cli

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"

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
//nolint:gocognit,funlen,forbidigo,cyclop,gocyclo,maintidx,nestif // TODO(Idelchi): Refactor.
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

	envFiles, err := expandFiles("env file", options.EnvFiles, logger)
	if err != nil {
		return err
	}

	if len(options.EnvFiles) > 0 && len(envFiles) == 0 {
		return fmt.Errorf("no env files matched the provided patterns: %v", options.EnvFiles)
	}

	envFromFiles := dotgen.Env{}
	activeEnvFiles := []string{}

	if len(envFiles) > 0 {
		vars, err := mergeVars(options, nil, "")
		if err != nil {
			return err
		}

		currentOS, ok := vars["OS"].(string)
		if !ok {
			return fmt.Errorf("expected string for OS, got %T", vars["OS"])
		}

		logger.Printlnf("processing %d env file(s)", len(envFiles))
		logger.Printlnf(" - processing:")

		for _, file := range envFiles {
			logger.Printlnf("  - %q", file)

			platformSuffix := getPlatformSuffixFromFileName(file)
			if !platformSuffixMatches(platformSuffix, currentOS) {
				logger.Printlnf(
					"    - skipping due to file suffix platform exclusion: file is for %q, current platform suffixes are %v",
					platformSuffix,
					activePlatformSuffixes(currentOS),
				)

				continue
			}

			activeEnvFiles = append(activeEnvFiles, file)
		}

		if len(activeEnvFiles) > 0 {
			for _, file := range activeEnvFiles {
				vars, err := mergeVars(options, nil, file)
				if err != nil {
					return err
				}

				data, err := os.ReadFile(filepath.Clean(file))
				if err != nil {
					return fmt.Errorf("loading env file: %w", err)
				}

				rendered, err := template.Apply(string(data), vars)
				if err != nil {
					return err //nolint:wrapcheck // Error is already descriptive enough.
				}

				env, err := godotenv.Unmarshal(rendered)
				if err != nil {
					return fmt.Errorf("parsing env file %q: %w", file, err)
				}

				maps.Copy(envFromFiles, env)
			}

			for key, value := range envFromFiles {
				if err := os.Setenv(key, value); err != nil {
					return fmt.Errorf("setting env %q: %w", key, err)
				}
			}
		}
	}

	files, err := expandFiles("config", options.Input, logger)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no files matched the provided patterns: %v", options.Input)
	}

	logger.Printlnf("processing %d file(s)", len(files))

	logger.Printlnf(" - processing:")

	included := make(map[string]string)

	for _, file := range activeEnvFiles {
		included[file] = envFromFiles.Export()
	}

	if len(envFromFiles) > 0 && !options.Dry && !options.Hash && !options.Debug {
		export, err := dotgen.Dotgen{Env: envFromFiles}.Export(options.Shell, "env files", false)
		if err != nil {
			return err //nolint:wrapcheck // Error is already descriptive enough.
		}

		if options.Verbose {
			printVerboseBlock(
				formatSources(activeEnvFiles),
				"Environment variables",
				format.Map(envFromFiles, "# %s=%q"),
			)
		}

		fmt.Println(export)
		fmt.Println()
	}

	for _, file := range files {
		logger.Printlnf("  - %q", file)

		vars, err := mergeVars(options, nil, file)
		if err != nil {
			return err
		}

		currentOS, ok := vars["OS"].(string)
		if !ok {
			return fmt.Errorf("expected string for OS, got %T", vars["OS"])
		}

		platformSuffix := getPlatformSuffixFromFileName(file)
		if !platformSuffixMatches(platformSuffix, currentOS) {
			logger.Printlnf(
				"    - skipping due to file suffix platform exclusion: file is for %q, current platform suffixes are %v",
				platformSuffix,
				activePlatformSuffixes(currentOS),
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
				return err //nolint:wrapcheck // Error is already descriptive enough.
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

			if header.Exclude.IsExcluded() {
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

		included[file] = vars.Export()

		if options.Dry || options.Hash {
			continue
		}

		vars.AppendCwd()

		rendered, err := template.Apply(string(doc), vars)
		if err != nil {
			return err //nolint:wrapcheck // Error is already descriptive enough.
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
			return err //nolint:wrapcheck // Error is already descriptive enough.
		}

		if err := dotgen.Validate(); err != nil {
			return err //nolint:wrapcheck // Error is already descriptive enough.
		}

		dotgen = dotgen.Filtered(currentOS, options.Shell)

		export, err := dotgen.Export(options.Shell, file, options.Instrument)
		if err != nil {
			return err //nolint:wrapcheck // Error is already descriptive enough.
		}

		if options.Verbose {
			printVerboseBlock(formatSources([]string{file}), "Template variables", format.Map(vars, "# %s=%q"))
		}

		fmt.Println(export)
		fmt.Println()
	}

	if options.Dry {
		for file := range included {
			fmt.Println(file)
		}

		return nil
	}

	if options.Hash {
		hash, err := format.Hash(included)
		if err != nil {
			return fmt.Errorf("computing hash: %w", err)
		}

		fmt.Print(hash)

		return nil
	}

	return nil
}
