package cli

import (
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"

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

	return platforms
}

// platformSuffixMatches checks whether a filename suffix applies to the current platform.
func platformSuffixMatches(suffix, operatingSystem string) bool {
	return suffix == "" || slices.Contains(activePlatformSuffixes(operatingSystem), suffix)
}
