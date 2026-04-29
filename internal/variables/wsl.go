package variables

import (
	"os"
	"runtime"
	"strings"
)

// IsWSL reports whether dotgen is running under Windows Subsystem for Linux.
func IsWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	for _, path := range []string{
		"/proc/sys/kernel/osrelease",
		"/proc/version",
	} {
		data, err := os.ReadFile(path) //nolint:gosec // Path is selected from a hardcoded list.
		if err != nil {
			continue
		}

		s := strings.ToLower(string(data))
		if strings.Contains(s, "microsoft") || strings.Contains(s, "wsl") {
			return true
		}
	}

	return false
}
