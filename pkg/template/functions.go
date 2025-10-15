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

// path returns the full path of a file by first checking if it exists in PATH;
// if not, checks if it exists as a full path to a file;
// otherwise returns an empty string.
func path(paths ...string) string {
	if len(paths) == 0 {
		return ""
	}

	path := filepath.ToSlash(filepath.Join(paths...))

	path, err := exec.LookPath(path)
	if err != nil {
		_, err := os.Stat(path)
		if err != nil {
			return ""
		}
	}

	return filepath.ToSlash(path)
}

// size returns the size of the file, if it does not exist of is a folder, returns 0.
func size(path string) int64 {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return 0
	}

	return info.Size()
}

// funcMap returns a map of custom template functions.
func funcMap() map[string]any {
	return map[string]any{
		"inPath":    inPath,
		"notInPath": func(name string) bool { return !inPath(name) },
		"exists":    exists,
		"path":      path,
		"size":      size,
	}
}
