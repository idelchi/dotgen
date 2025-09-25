#!/bin/sh
set -e

# Tool specific variables
TOOL="aliaser"
DISABLE_SSL="${ALIASER_DISABLE_SSL:-false}"

need_cmd() {
  if ! command -v "${1}" >/dev/null 2>&1; then
    printf "Required command '%s' not found\n" "$1"
    exit 1
  fi
}

main() {
  # Check for required commands
  need_cmd curl

  # Call the installation script with the provided arguments
  # shellcheck disable=SC2312
  curl ${DISABLE_SSL:+-k} -sSL https://raw.githubusercontent.com/idelchi/scripts/refs/heads/main/install.sh | INSTALLER_TOOL=${TOOL} sh -s -- "$@"
}

main "$@"
