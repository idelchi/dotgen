package template

import (
	"os"
	"os/exec"
	"path/filepath"
)

// inPath checks whether a command is available in PATH.
func inPath(name string) bool {
	_, err := exec.LookPath(name)

	return err == nil
}

// exists checks whether a file or directory exists at the given path.
func exists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

// path returns the full path of a command if it exists in PATH, otherwise returns an empty string.
func path(name string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}

	return filepath.ToSlash(path)
}

// funcMap returns a map of custom template functions.
func funcMap() map[string]any {
	return map[string]any{
		"inPath":    inPath,
		"notInPath": func(name string) bool { return !inPath(name) },
		"exists":    exists,
		"path":      path,
	}
}
