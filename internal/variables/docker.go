package variables

import (
	"os"
	"runtime"
)

// IsDocker reports whether dotgen is running inside a Docker container.
func IsDocker() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	_, err := os.Stat("/.dockerenv")

	return err == nil
}
