package cli

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/idelchi/dotgen/internal/variables"
)

// ErrExitGracefully is used to signal a graceful exit without error.
var ErrExitGracefully = errors.New("exit gracefully")

// Variables are layered as:
//
//	Defaults (from variables.Defaults)
//	Values from the header of the config file (if any)
//	Values files (from --values)
//	Command-line args (from key=value pairs)
func mergeVars(options Options, headers variables.Variables, file string) (variables.Variables, error) {
	// Load default variables first
	vars := variables.Defaults(options.Shell, file)

	if headers != nil {
		maps.Copy(vars, headers)
	}

	values, err := variables.Values(options.Values).Variables()
	if err != nil {
		return nil, err //nolint:wrapcheck // Error is already descriptive enough.
	}

	maps.Copy(vars, values)

	args, err := variables.Args(options.Set).ToKeyValues()
	if err != nil {
		return nil, fmt.Errorf("parsing args: %w", err)
	}

	maps.Copy(vars, args)

	return vars, nil
}

// normalizePatterns expands directory-like patterns to use the given default path.
// It returns forward-slash-normalized patterns.
func normalizePatterns(patterns []string, defaultPath string) []string {
	for idx, pattern := range patterns {
		pattern = filepath.ToSlash(pattern)

		switch {
		case pattern == ".":
			pattern = defaultPath
		case strings.HasSuffix(pattern, "/"):
			pattern = filepath.Join(pattern, defaultPath)
		default:
			info, err := os.Stat(pattern)
			if err == nil && info.IsDir() {
				pattern = filepath.Join(pattern, defaultPath)
			}
		}

		patterns[idx] = filepath.ToSlash(pattern)
	}

	return patterns
}

// expandFiles expands glob patterns into file paths.
// It returns forward-slash-normalized paths for files matching the provided patterns.
func expandFiles(kind string, patterns []string, logger Logger) ([]string, error) {
	files := []string{}

	for _, file := range patterns {
		file = filepath.ToSlash(file)

		base, pattern := doublestar.SplitPattern(file)
		fsys := os.DirFS(base)

		logger.Printlnf("loading %s %q", kind, file)

		matches, err := doublestar.Glob(fsys, pattern, doublestar.WithFilesOnly())
		if err != nil {
			return nil, fmt.Errorf("invalid %s pattern %q: %w", kind, file, err)
		}

		logger.Printlnf(" - found %d file(s) for pattern %q", len(matches), file)

		for _, file := range matches {
			file = filepath.ToSlash(filepath.Join(base, file))
			files = append(files, file)
		}
	}

	return files, nil
}

// formatSources formats file paths for a generated source comment.
func formatSources(files []string) string {
	sources := make([]string, 0, len(files))
	for _, file := range files {
		sources = append(sources, fmt.Sprintf("%q", file))
	}

	return strings.Join(sources, ", ")
}

// printVerboseBlock prints the verbose generated output header and metadata block.
//
//nolint:forbidigo // Function needs to print to the console directly.
func printVerboseBlock(source, label, values string) {
	line := fmt.Sprintf("# Generated from %s", source)
	date := fmt.Sprintf("# Date: %s", time.Now().Format(time.RFC3339))
	stars := strings.Repeat("*", len(line))

	fmt.Println("# " + stars)
	fmt.Println(line)
	fmt.Println(date)
	fmt.Println("# " + stars)

	fmt.Printf("# %s:\n", label)
	fmt.Println("# " + stars)
	fmt.Println(values)
	fmt.Println("# " + stars)
	fmt.Println()
}

// getPlatformSuffixFromFileName checks if the file name ends with _<platform> before the extension.
// It returns the platform suffix if found, otherwise an empty string.
func getPlatformSuffixFromFileName(file string) string {
	base := filepath.Base(file)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	parts := strings.Split(name, "_")

	const expectedParts = 2

	if len(parts) < expectedParts {
		return ""
	}

	platform := parts[len(parts)-1]
	knownPlatforms := []string{
		"linux",
		"darwin",
		"windows",
		"freebsd",
		"openbsd",
		"netbsd",
		"dragonfly",
		"solaris",
		"aix",
		"wsl",
		"docker",
	}

	if slices.Contains(knownPlatforms, platform) {
		return platform
	}

	return ""
}

// activePlatformSuffixes returns the filename suffixes that match the current platform.
func activePlatformSuffixes(operatingSystem string) []string {
	platforms := []string{operatingSystem}

	if variables.IsWSL() {
		platforms = append(platforms, "wsl")
	}

	if variables.IsDocker() {
		platforms = append(platforms, "docker")
	}

	return platforms
}

// platformSuffixMatches checks whether a filename suffix applies to the current platform.
func platformSuffixMatches(suffix, operatingSystem string) bool {
	return suffix == "" || slices.Contains(activePlatformSuffixes(operatingSystem), suffix)
}
