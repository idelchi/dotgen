// Package variables provides functionality for managing key-value pairs and loading them from files or command-line
// arguments.
package variables

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v4/host"

	"github.com/idelchi/dotgen/internal/format"

	"go.yaml.in/yaml/v4"
)

// Variables represents a set of key-value pairs.
type Variables map[string]any

// Export returns a string representation of the variables, as a comment block.
func (v Variables) Export() string {
	return format.Map(v, "%s=%q")
}

// Values represents a list of value file paths.
type Values []string

// Variables loads and merges variables from the specified value files.
func (v Values) Variables() (Variables, error) {
	variables := make(Variables)

	for _, path := range v {
		data, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return nil, fmt.Errorf("loading values file: %w", err)
		}

		values := make(Variables)
		if err := yaml.Unmarshal(data, &values); err != nil {
			return nil, fmt.Errorf("parsing values file: %w", err)
		}

		maps.Copy(variables, values)
	}

	return variables, nil
}

// Defaults returns a set of default variables based on the current environment.
func Defaults(shell, file string) Variables {
	variables := make(Variables)

	variables["OS"] = runtime.GOOS

	if file != "" {
		variables["DOTGEN_CURRENT_FILE"] = filepath.ToSlash(file)
		variables["DOTGEN_CURRENT_DIR"] = filepath.ToSlash(filepath.Dir(file))
	}

	info, err := host.Info()
	if err == nil {
		variables["HOSTNAME"] = info.Hostname
		variables["PLATFORM"] = info.Platform
		variables["ARCHITECTURE"] = info.KernelArch
	}

	variables["USER"] = os.Getenv("USER")
	variables["USERNAME"] = os.Getenv("USERNAME")
	variables["HOME"] = filepath.ToSlash(os.Getenv("HOME"))

	if usr, err := user.Current(); err == nil {
		if variables["USER"] == "" {
			variables["USER"] = usr.Name
		}

		if variables["USERNAME"] == "" {
			variables["USERNAME"] = usr.Username
		}

		if variables["HOME"] == "" {
			variables["HOME"] = filepath.ToSlash(usr.HomeDir)
		}
	}

	//nolint:errcheck,forcetypeassert // HOME cannot be non-string here.
	home := variables["HOME"].(string)

	variables["CONFIG_DIR"] = filepath.ToSlash(filepath.Join(home, ".config"))
	variables["CACHE_DIR"] = filepath.ToSlash(filepath.Join(home, ".cache"))
	variables["TMP_DIR"] = filepath.ToSlash(os.TempDir())

	variables["SHELL"] = shell

	variables["EXTENSION"] = ""

	if runtime.GOOS == "windows" {
		variables["EXTENSION"] = ".exe"
	}

	return variables
}

// AppendCwd appends the current working directory to the variables under the "CWD" key.
func (v *Variables) AppendCwd() {
	cwd, err := os.Getwd()
	if err != nil {
		cwd, err = filepath.Abs(".")
		if err != nil {
			cwd = "."
		}
	}

	(*v)["CWD"] = filepath.ToSlash(cwd)
}

// Args represents a list of key=value strings.
type Args []string

// ToKeyValues converts the Args into a Variables map, returning any errors encountered.
func (a Args) ToKeyValues() (Variables, error) {
	var errs []error

	out := make(Variables)

	for _, keyValue := range a {
		key, value, found := strings.Cut(keyValue, "=")

		if !found {
			errs = append(errs, fmt.Errorf("missing value for %q", keyValue))

			continue
		}

		if key == "" {
			errs = append(errs, fmt.Errorf("missing key for %q", keyValue))

			continue
		}

		out[key] = value
	}

	return out, errors.Join(errs...)
}
