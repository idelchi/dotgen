package dotgen

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/MakeNowJust/heredoc/v2"
)

// ToShellVar converts an arbitrary string into a valid shell variable name.
func ToShellVar(s string) string {
	// Replace invalid characters with underscores
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	safe := re.ReplaceAllString(s, "_")

	// If first char isn’t letter or underscore, prepend underscore
	if len(safe) == 0 || (!unicode.IsLetter(rune(safe[0])) && safe[0] != '_') {
		safe = "_" + safe
	}

	return strings.TrimRight(safe, "\n")
}

// instrumentationSummary returns an awk script that summarizes the
// instrumentation results stored in the given variable.
func instrumentationSummary(varName string) string {
	awkScript := heredoc.Doc(`
		if command -v awk >/dev/null 2>&1; then
		  LC_ALL=C printf '%s\n' "${__VAR__[@]}" \
		  | sort -k2,2nr -k1,1 \
		  | awk '{
		      printf("(%4d ms) %s\n", $2+0, $1)
		      total += $2
		    }
		    END {
		      printf("\nTotal: %d ms\n", total)
		    }'
		else
		  printf '%s\n' "${__VAR__[@]}"
		fi
	`)

	return strings.ReplaceAll(awkScript, "__VAR__", varName)
}
