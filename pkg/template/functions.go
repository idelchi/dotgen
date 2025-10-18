package template

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/idelchi/dotgen/internal/format"
)

// _which returns the full path of an executable if it exists in PATH,
// otherwise returns an empty string along with an error.
func _which(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", err //nolint:wrapcheck	// Error is already descriptive enough
	}

	return filepath.ToSlash(path), nil
}

// which returns the full path of an executable if it exists in PATH,
// otherwise returns an empty string.
func which(name string) string {
	path, err := _which(name)
	if err != nil {
		return ""
	}

	return path
}

// inPath checks whether a command is available in PATH.
func inPath(name string) bool {
	_, err := _which(name)

	return err == nil
}

// notInPath checks whether a command is not available in PATH.
func notInPath(name string) bool {
	_, err := _which(name)

	return err != nil
}

// exists checks whether a file or directory exists at the given path.
func exists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

// resolve returns the full path of a file if it exists,
// otherwise returns an empty string.
func resolve(paths ...string) string {
	if len(paths) == 0 {
		return ""
	}

	path := filepath.ToSlash(filepath.Join(paths...))

	if _, err := os.Stat(path); err != nil {
		return ""
	}

	return path
}

// size returns the size of the file, if it does not exist of is a folder, returns 0.
func size(path string) int64 {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return 0
	}

	return info.Size()
}

// join joins multiple path elements into a single path.
func join(paths ...string) string {
	return filepath.ToSlash(filepath.Join(paths...))
}

// read attempts to read the file at the location and returns its content as a string.
func read(path string) (string, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return "", err //nolint:wrapcheck	// Error is already descriptive enough
	}

	return string(data), nil
}

// copy copies a file from src to dst, creating the destination directory if needed.
// Returns an error if the source doesn't exist, isn't a regular file, or if the copy fails.
func copyFile(src, dst string) (string, error) {
	data, err := os.ReadFile(filepath.Clean(src))
	if err != nil {
		return "", err //nolint:wrapcheck	// Error is already descriptive enough
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return "", err //nolint:wrapcheck	// Error is already descriptive enough
	}

	if err := os.WriteFile(filepath.Clean(dst), data, 0o600); err != nil {
		return "", err //nolint:wrapcheck	// Error is already descriptive enough
	}

	return "", nil
}

// posixPath converts a Windows path (like `C:/...`) to WSL format (`/c/...`).
// On non-Windows systems, it returns the path unchanged.
func posixPath(path string) string {
	return format.PosixPath(path)
}

// windowsPath converts a Windows path (like `C:/...`) to WSL format (`/c/...`).
// On non-Windows systems, it returns the path unchanged.
func windowsPath(path string) string {
	return format.WindowsPath(path)
}

// FuncMap returns a map of custom template functions.
func FuncMap() map[string]any {
	return map[string]any{
		"inPath":      inPath,
		"notInPath":   notInPath,
		"exists":      exists,
		"which":       which,
		"resolve":     resolve,
		"size":        size,
		"join":        join,
		"read":        read,
		"copy":        copyFile,
		"posixPath":   posixPath,
		"windowsPath": windowsPath,
	}
}
