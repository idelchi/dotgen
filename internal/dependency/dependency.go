// Package dependency provides deterministic fingerprints for external configuration inputs.
package dependency

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/idelchi/dotgen/internal/format"
)

// Dependencies represents external files and executables that affect generated output.
type Dependencies struct {
	// Files contains file paths or glob patterns relative to the declaring configuration file.
	Files []string `yaml:"files,omitempty"`
	// Executables contains command names to resolve from PATH.
	Executables []string `yaml:"executables,omitempty"`
}

// Fingerprints resolves the dependencies and returns deterministic records for hashing and diagnostics.
func (d Dependencies) Fingerprints(baseDir string) ([]string, error) {
	records, err := fingerprintFiles(d.Files, baseDir)
	if err != nil {
		return nil, err
	}

	executables, err := fingerprintExecutables(d.Executables)
	if err != nil {
		return nil, err
	}

	records = append(records, executables...)
	sort.Strings(records)

	return records, nil
}

// fingerprintFiles expands file patterns and returns their content fingerprints.
func fingerprintFiles(patterns []string, baseDir string) ([]string, error) {
	records := []string{}
	seen := map[string]struct{}{}

	for _, pattern := range patterns {
		if !filepath.IsAbs(pattern) {
			pattern = filepath.Join(baseDir, pattern)
		}

		pattern = filepath.ToSlash(pattern)

		base, glob := doublestar.SplitPattern(pattern)

		matches, err := doublestar.Glob(os.DirFS(base), glob, doublestar.WithFilesOnly())
		if err != nil {
			return nil, fmt.Errorf("invalid dependency file pattern %q: %w", pattern, err)
		}

		if len(matches) == 0 {
			records = append(records, fmt.Sprintf("file pattern %q: no matches", pattern))

			continue
		}

		for _, match := range matches {
			path, err := absolutePath(filepath.Join(base, match))
			if err != nil {
				return nil, err
			}

			if _, ok := seen[path]; ok {
				continue
			}

			digest, err := format.Fingerprint(path)
			if err != nil {
				return nil, fmt.Errorf("fingerprinting dependency file %q: %w", path, err)
			}

			records = append(records, fmt.Sprintf("file %q: sha256=%s", path, digest))
			seen[path] = struct{}{}
		}
	}

	return records, nil
}

// fingerprintExecutables resolves executable names and returns their filesystem metadata.
func fingerprintExecutables(names []string) ([]string, error) {
	records := make([]string, 0, len(names))
	seen := map[string]struct{}{}

	for _, name := range names {
		if _, ok := seen[name]; ok {
			continue
		}

		seen[name] = struct{}{}

		lookupPath, err := exec.LookPath(name)
		if err != nil {
			records = append(records, fmt.Sprintf("executable %q: missing", name))

			continue
		}

		lookupPath, err = absolutePath(lookupPath)
		if err != nil {
			return nil, fmt.Errorf("resolving executable %q: %w", name, err)
		}

		targetPath, err := filepath.EvalSymlinks(lookupPath)
		if err != nil {
			return nil, fmt.Errorf("resolving executable %q target: %w", name, err)
		}

		targetPath, err = absolutePath(targetPath)
		if err != nil {
			return nil, fmt.Errorf("resolving executable %q target: %w", name, err)
		}

		info, err := os.Stat(targetPath)
		if err != nil {
			return nil, fmt.Errorf("inspecting executable %q: %w", name, err)
		}

		records = append(records, fmt.Sprintf(
			"executable %q: path=%q target=%q size=%d modified=%d",
			name,
			lookupPath,
			targetPath,
			info.Size(),
			info.ModTime().UnixNano(),
		))
	}

	return records, nil
}

// absolutePath returns a forward-slash-normalized absolute path.
func absolutePath(path string) (string, error) {
	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("resolving path %q: %w", path, err)
	}

	return filepath.ToSlash(absolute), nil
}
