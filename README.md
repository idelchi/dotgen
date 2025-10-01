# dotgen

A tool to manage and render dotfiles.

---

[![GitHub release](https://img.shields.io/github/v/release/idelchi/dotgen)](https://github.com/idelchi/dotgen/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/idelchi/dotgen.svg)](https://pkg.go.dev/github.com/idelchi/dotgen)
[![Go Report Card](https://goreportcard.com/badge/github.com/idelchi/dotgen)](https://goreportcard.com/report/github.com/idelchi/dotgen)
[![Build Status](https://github.com/idelchi/dotgen/actions/workflows/github-actions.yml/badge.svg)](https://github.com/idelchi/dotgen/actions/workflows/github-actions.yml/badge.svg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`dotgen` is a command-line utility for defining, templating, and exporting dotfiles.

Allows for defining dotfiles as YAML, applying Go template substitution and exporting as shell-ready scripts.

- `env` variables, rendered as `export KEY=VALUE`
- `vars` definitions, rendered as `KEY=VALUE`
- `alias` and `function` definitions,
- `raw` shell snippets
- `run` commands have render the stdout in-place or or exported to files to be sourced

Example use case is to add the following to your `.rc` file (e.g. `.zshrc`):

```sh
if [ ! -d "${HOME}/.cache/dotgen" ]; then
  mkdir -p "${HOME}/.cache/dotgen"
  dotgen -i "/path/to/configs/**/*.dotgen" --shell zsh > "${HOME}/.cache/dotgen/dotgen.rc"
fi

source "$HOME/.cache/dotgen/dotgen.rc"
```

or simply

```sh
eval "$(dotgen -i "/path/to/configs/**/*.dotgen" --shell zsh)"
```

## Installation

```sh
curl -sSL https://raw.githubusercontent.com/idelchi/dotgen/refs/heads/main/install.sh | sh -s -- -d ~/.local/bin
```

## Usage

```sh
# Run dotgen with a config file
$ dotgen -i ~/.config/dotgen/config.dotgen --shell zsh
```

```sh
# Provide additional variables inline for use in templates
$ dotgen -i config.dotgen ENABLE_FZF_FEATURES=true KEY=VALUE
```

```sh
# Merge extra variables from YAML value files
$ dotgen -i config.dotgen -V values.yml
```

### Configuration

The configuration file consists of a header section where you can define values to be used in templates, as well
as a global exclude condition.

The second part of the file contains the environment variables, variables, and commands which
will be parsed and exported.

The header section can be omitted if not used.

<!-- prettier-ignore-start -->
```yaml
values:
  BIN_DIR: $HOME/bin

---
env:
  EDITOR: nano

vars:
  project: /work/src/myproject

commands:
  - name: gs
    cmd: git status
    kind: alias
    exclude: {{ notInPath "git" }}

  - name: hostname
    doc: export hostname as variable
    kind: run
    cmd: |
      echo "export HOSTNAME=$(hostname -f 2>/dev/null || hostname)"

  - name: zoxide integration
    doc: zoxide shell integration
    kind: run
    cmd: zoxide init "{{ .SHELL }}" --cmd cd
    export_to: {{ .CACHE_DIR }}/01-zoxide.rc
    exclude: {{notInPath "zoxide"}}

  - name: starship integration
    doc: starship shell integration
    kind: run
    cmd: starship init --print-full-init "{{ .SHELL }}"
    export_to: {{ .CACHE_DIR }}/00-starship.rc
    exclude: {{ notInPath "starship" }}

  - name: greet
    kind: function
    cmd: |
      echo "Hello from '{{ .OS }}' on '{{ .ARCHITECTURE }}'!"
      echo "Project base is ${project}"
```

will be rendered as:

```sh
# Environment variables
# ------------------------------------------------
export EDITOR="nano"
# ------------------------------------------------

# Variables
# ------------------------------------------------
project="/work/src/myproject"
# ------------------------------------------------

# Commands
# ------------------------------------------------
# name: gs
alias gs='git status'

# name: hostname
# doc:
#  export hostname as variable
# original:
#  echo "export HOSTNAME=$(hostname -f 2>/dev/null || hostname)"
export HOSTNAME=dotgen

# name: starship integration
# doc:
#  starship shell integration
# original:
#  starship init --print-full-init "zsh"
# output exported to "/home/user/.cache/00-starship.rc"
. "/home/user/.cache/00-starship.rc"

# name: greet
greet() {
  echo "Hello from 'linux' on 'x86_64'!"
  echo "Project base is ${project}"
}

# ------------------------------------------------
```
<!-- prettier-ignore-end -->

Run with `--verbose` to embed more details in the output.

### Command Kinds

- **alias** — Simple shell alias
- **function** — Shell function
- **raw** — Raw shell snippet, rendered as-is
- **run** — Command to be executed, stdout is captured and rendered as-is or exported

### Filtering

Commands can be limited to specific OS or shell:

```yaml
commands:
  - name: ll
    cmd: ls -al
    kind: alias
    os:
      - linux
      - darwin
    shell:
      - zsh
      - bash
```

### Excluding

Commands can be marked to be exclude:

<!-- prettier-ignore-start -->
```yaml
commands:
  - name: debug
    cmd: echo "debug only"
    exclude: {{ eq .HOSTNAME "something" }}
```
<!-- prettier-ignore-end -->

`dotgen` will also skip files on the form `<name>_<os>.<extension>` if the current value of `OS` does not match `<os>`.

## Variables

Variables are populated by default as:

- Platform (`OS`, `PLATFORM`, `ARCHITECTURE`, `HOSTNAME`)
- User (`USER`, `USERNAME`, `HOME`, `CACHE_DIR`, `CONFIG_DIR`, `TMPDIR`)
- Active shell (`SHELL`)

You can add more through:

- Inline args: `KEY=VALUE`
- Value files via `-V file.yml`
- Header section in the config file

Merge order is:

1. Default variables
2. Header section in the config file
3. Value files via `-V file.yml`
4. Inline args: `KEY=VALUE`

## Templating

All config files are processed as Go templates, extended with [slim-sprig](https://go-task.github.io/slim-sprig) functions.

Additional custom functions:

- `inPath "cmd"` - Check if a command is available in PATH
- `notInPath "cmd"` - Check if a command is not available in PATH
- `exists "path"` - Check if a file or folder exists

## Flags

- `--input`, `-i` - File paths or patterns (`doublestar`) to process
- `--shell`, `-s` - Active shell name (default: `$SHELL` or `zsh`)
- `--values`, `-V` - Additional YAML values to merge and use in templates
- `--verbose` - Enable verbose logging
- `--debug`, `-d` - Show variables and rendered templates without further processing. Implies `--verbose`
- `--instrument`, `-I` - Instrument command execution times and show summary.
  `$?` will not be respected between commands. Implies `--verbose`
- `--version`, `-v` - Show version

## Demo

![Demo](assets/gifs/dotgen.gif)
