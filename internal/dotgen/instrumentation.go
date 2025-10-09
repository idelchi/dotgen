package dotgen

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/MakeNowJust/heredoc/v2"
)

// Instrumentation represents the instrumentation configuration for a Dotgen file.
type Instrumentation struct {
	// Name is the name of the Dotgen file being instrumented.
	Name string
	// Variable is the name of the shell variable used to store instrumentation data.
	variable string

	// Enabled indicates whether instrumentation is enabled.
	Enabled bool
}

// Instrument creates an Instrumentation instance for the given Dotgen file.
func Instrument(file string) Instrumentation {
	variable := fmt.Sprintf("__dotgen_instrumentation_%s", toShellVar(file))

	return Instrumentation{Name: file, variable: variable, Enabled: true}
}

// Disable disables the instrumentation.
func (i *Instrumentation) Disable() {
	i.Enabled = false
}

// Header returns the header section for instrumentation.
func (i Instrumentation) Header() string {
	if !i.Enabled {
		return ""
	}

	return heredoc.Docf(`
		# Instrumentation
		# ------------------------------------------------
		%s=()
		# ------------------------------------------------
	`, i.variable)
}

// Wrap wraps a command with instrumentation code to measure its execution time.
func (i Instrumentation) Wrap(name, command string) string {
	if !i.Enabled {
		return command
	}

	shellName := toShellVar(name)

	prefix := fmt.Sprintf("__dotgen_%s_start", shellName)
	suffix := fmt.Sprintf("__dotgen_%s_end", shellName)
	elapsed := fmt.Sprintf("__dotgen_%s_elapsed", shellName)

	return heredoc.Docf(`

		# Instrumentation for: %s
		# ------------------------------------------------

		%s=$(date +%%s%%3N)

		# Command to measure
		# ------------------------------------------------
		%s
		# ------------------------------------------------

		%s=$(date +%%s%%3N)
		%s=$((%s - %s))
		%s+=("%s ${%s}")

	`, name, prefix, strings.TrimRight(command, "\n"), suffix, elapsed, suffix, prefix, i.variable, shellName, elapsed)
}

// Footer returns the footer section for instrumentation, including a summary of execution times.
func (i Instrumentation) Footer() string {
	if !i.Enabled {
		return ""
	}

	return heredoc.Docf(`
		echo '************************************************'
		echo "[dotgen instrumentation] summary for %s:"
		if command -v awk >/dev/null 2>&1; then
		  LC_ALL=C printf '%%s\n' "${%s[@]}" \
		  | sort -k2,2nr -k1,1 \
		  | awk '{
		      printf("(%%4d ms) %%s\n", $2+0, $1)
		      total += $2
		    }
		    END {
		      printf("\nTotal: %%d ms\n", total)
		    }'
		else
		  printf '%%s\n' "${%s[@]}"
		fi
		echo '************************************************'
	`, i.Name, i.variable, i.variable)
}

// toShellVar converts an arbitrary string into a valid shell variable name.
func toShellVar(s string) string {
	// Replace invalid characters with underscores
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	safe := re.ReplaceAllString(s, "_")

	// If first char isn’t letter or underscore, prepend underscore
	if len(safe) == 0 || (!unicode.IsLetter(rune(safe[0])) && safe[0] != '_') {
		safe = "_" + safe
	}

	return strings.TrimRight(safe, "\n")
}
