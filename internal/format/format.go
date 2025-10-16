// Package format provides functions to format shell scripts and file paths.
package format

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"mvdan.cc/sh/v3/syntax"
)

// Shell formats a shell script source code.
func Shell(src string, singleLine bool) (string, error) {
	file, err := syntax.NewParser(syntax.KeepComments(true)).Parse(strings.NewReader(src), "")
	if err != nil {
		return "", err //nolint:wrapcheck	// TODO(Idelchi): Inspect if we want to wrap this error.
	}

	const indent = 2

	options := []syntax.PrinterOption{
		syntax.Indent(indent),
	}

	if singleLine {
		options = append(options, syntax.SingleLine(true))
	}

	printer := syntax.NewPrinter(options...)

	var buf bytes.Buffer
	if err := printer.Print(&buf, file); err != nil {
		return "", err //nolint:wrapcheck	// TODO(Idelchi): Inspect if we want to wrap this error.
	}

	return buf.String(), nil
}

// isAlpha checks if a byte is an ASCII letter.
func isAlpha(b byte) bool {
	return unicode.IsLetter(rune(b))
}

// WindowsPath formats a file path to use forward slashes and cleans it.
// Converts paths like `/c/...` to `C:/...` regardless of platform if a drive letter is detected.
func WindowsPath(path string) string {
	path = filepath.ToSlash(path)

	if len(path) >= 3 && path[0] == '/' && path[2] == '/' && isAlpha(path[1]) {
		drive := strings.ToUpper(path[1:2])

		path = drive + ":" + path[2:]
	}

	// Convert slashes to forward slashes.
	path = filepath.ToSlash(path)

	return path
}

// PosixPath converts a Windows path (like `C:/...`) to Posix format (`/c/...`).
// Converts regardless of platform if a drive letter is detected.
func PosixPath(path string) string {
	path = filepath.ToSlash(path)

	// Check for drive letter pattern: X:/ where X is a letter
	if len(path) >= 3 && path[1] == ':' && path[2] == '/' && isAlpha(path[0]) {
		drive := strings.ToLower(path[0:1])

		path = "/" + drive + path[2:]
	}

	// Convert slashes to forward slashes.
	path = filepath.ToSlash(path)

	return path
}

// Map takes a map and a format string (like "export %s=%q")
// and returns a sorted, joined string.
func Map[T any](data map[string]T, format string) string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, fmt.Sprintf(format, k, data[k]))
	}

	return strings.Join(out, "\n")
}

// Hash deterministically combines each file's contents and its exported variables into a single SHA-256 digest.
func Hash(included map[string]string) (string, error) {
	if len(included) == 0 {
		return "", errors.New("no input provided")
	}

	digests := make([]string, 0, len(included))

	for name, vars := range included {
		data, err := os.ReadFile(filepath.Clean(name))
		if err != nil {
			return "", fmt.Errorf("reading %q: %w", name, err)
		}

		fileHash := sha256.Sum256(data)
		exportHash := sha256.Sum256([]byte(vars))

		entry := name + "\n" +
			hex.EncodeToString(fileHash[:]) + "\n" +
			hex.EncodeToString(exportHash[:]) + "\n"

		entryHash := sha256.Sum256([]byte(entry))

		digests = append(digests, hex.EncodeToString(entryHash[:]))
	}

	sort.Strings(digests)

	payload := strings.Join(digests, "\n") + "\n"
	final := sha256.Sum256([]byte(payload))

	return hex.EncodeToString(final[:]), nil
}
