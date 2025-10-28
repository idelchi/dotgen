# dotgen

A tool to manage and render dotfiles.

---

[![GitHub release](https://img.shields.io/github/v/release/idelchi/dotgen)](https://github.com/idelchi/dotgen/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/idelchi/dotgen.svg)](https://pkg.go.dev/github.com/idelchi/dotgen)
[![Go Report Card](https://goreportcard.com/badge/github.com/idelchi/dotgen)](https://goreportcard.com/report/github.com/idelchi/dotgen)
[![Build Status](https://github.com/idelchi/dotgen/actions/workflows/github-actions.yml/badge.svg)](https://github.com/idelchi/dotgen/actions/workflows/github-actions.yml/badge.svg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Define your dotfiles once as YAML, template them with Go, and render them for any OS or shell.

## What it does

`dotgen` takes YAML files like this:

<!-- prettier-ignore-start -->
```yaml
env:
  EDITOR: nano

vars:
  project: /work/myproject

commands:
  - name: gs
    cmd: git status
    kind: alias
    exclude: {{ notInPath "git" }}

  - name: greet
    kind: function
    cmd: |
      echo "Hello from {{ .OS }} on {{ .ARCHITECTURE }}!"
      echo "Project at ${project}"
```
<!-- prettier-ignore-end -->

And outputs shell code you can source:

```sh
# Environment variables
# ------------------------------------------------
export EDITOR="nano"
# ------------------------------------------------

# Variables
# ------------------------------------------------
project="/work/myproject"
# ------------------------------------------------

# Commands
# ------------------------------------------------
# name: gs
alias gs='git status'

# name: greet
greet() {
  echo "Hello from linux on x86_64!"
  echo "Project at ${project}"
}

# ------------------------------------------------
```

Template with Go, filter by OS/shell, conditionally exclude commands,
run setup scripts once - all from declarative config files.

## Installation

```sh
curl -sSL https://raw.githubusercontent.com/idelchi/dotgen/refs/heads/main/install.sh | sh -s -- -d ~/.local/bin
```

## Quick start

```sh
# Generate and source your dotfiles
eval "$(dotgen --shell zsh ~/.config/dotgen/**/*.dotgen)"
```

or cache the output to avoid regenerating on every shell startup:

```sh
# In your .zshrc or .bashrc
if [ ! -f "${HOME}/.cache/dotgen/dotgen.rc" ]; then
  mkdir -p "${HOME}/.cache/dotgen"
  dotgen --shell zsh "/path/to/configs/**/*.dotgen" > "${HOME}/.cache/dotgen/dotgen.rc"
fi

source "${HOME}/.cache/dotgen/dotgen.rc"
```

## Configuration

Config files are YAML with an optional header section for template variables and
a body section for your actual dotfiles.

<!-- prettier-ignore-start -->
```yaml
# Header (optional) - define template variables and global excludes
values:
  BIN_DIR: ${HOME}/bin

exclude: {{ notInPath "git" }}

---
# Body - your actual dotfiles
env:
  PATH: "{{ .BIN_DIR }}:${PATH}"
  EDITOR: nano

vars:
  project: /work/src/myproject

commands:
  - name: gs
    cmd: git status
    kind: alias

  - name: greet
    kind: function
    cmd: echo "Hello!"
```
<!-- prettier-ignore-end -->

### Command types

Four command kinds, each with different output behavior:

**`alias`** - Shell alias

```yaml
- name: ll
  cmd: ls -al
  # default, can be omitted
  kind: alias
```

```sh
alias ll='ls -al'
```

**`function`** - Shell function

```yaml
- name: greet
  kind: function
  cmd: |
    echo "Hello $1"
```

```sh
greet() {
  echo "Hello $1"
}
```

**`raw`** - Raw shell code, no wrapping

```yaml
- name: custom
  kind: raw
  cmd: |
    # Direct shell code
    if [ -f ${HOME}/.rc ]; then
      source ${HOME}/.rc
    fi
```

**`run`** - Execute command at generation time, capture stdout

```yaml
- name: zoxide
  kind: run
  cmd: zoxide init "{{ .SHELL }}" --cmd cd
  export_to: {{ .CACHE_DIR }}/zoxide.rc
  timeout: 30s
```

Use `run` for tool integrations (starship, zoxide, etc.) that need to execute once to generate shell code.

`export_to` takes three possible values:

- `""`, `null` or omitted: stdout (inserted in place)
- `path`: writes output to that file, and inserts a `. <path>` line instead
- `/dev/null`: discarded. Allows for just performing some operations without outputting anything.

`timeout` accepts Go duration format (e.g., `30s`, `5m`, `1h30m`). Defaults to `1m` if not specified.

### Filtering

Target specific operating systems or shells:

```yaml
commands:
  - name: pbcopy-alias
    cmd: xclip -selection clipboard
    kind: alias
    os:
      - linux
    shell:
      - zsh
      - bash
```

Or use template logic to exclude conditionally:

<!-- prettier-ignore-start -->
```yaml
commands:
  - name: gs
    cmd: git status
    kind: alias
    exclude: {{ notInPath "git" }}
```
<!-- prettier-ignore-end -->

Files named `config_<os>.dotgen` are automatically skipped if the OS doesn't match.

## Variables

Every template has access to these built-in variables:

- **Platform**: `OS`, `PLATFORM`, `ARCHITECTURE`, `EXTENSION`, `HOSTNAME`
- **User**: `USER`, `USERNAME`, `HOME`, `CACHE_DIR`, `CONFIG_DIR`, `TMPDIR`
- **Shell**: `SHELL`
- **File context**: `DOTGEN_CURRENT_FILE`, `DOTGEN_CURRENT_DIR`

Add your own variables in multiple ways:

```sh
# Inline
dotgen config.dotgen --set KEY=VALUE --set ANOTHER=thing

# From YAML files
dotgen config.dotgen --values values.yml

# In config header
values:
  MY_VAR: some_value
```

Variables merge in this order (last wins):

1. Built-in defaults
2. Config file header
3. `--values` value files
4. `--set KEY=VALUE` args

## Templating

All config files are processed as Go templates with [slim-sprig](https://go-task.github.io/slim-sprig)
functions plus custom helpers:

- `inPath "cmd"` - Check if command exists in PATH
- `notInPath "cmd"` - Inverse of above
- `exists "path"` - Check if file/directory exists
- `which "cmd"` - Get full path to a command if in PATH, empty string if not found
- `resolve "paths"...` - Joins multiple path elements and returns the full path if it exists, empty string otherwise
- `size "path"` - Get file size in bytes, 0 if missing
- `join "paths"...` - Join multiple path elements into a single path
- `read "path"` - Read file content, returns an error if the file doesn't exist or can't be read
- `posixPath "path"` - Convert Windows path (like `C:/...` or `C:\...`) to Posix format (`/c/...`)
- `windowsPath "path"` - Convert Posix path (like `/c/...`) to Windows format (`C:/...`)
- `mustEnv "KEY"` - Return the value of an environment variable, or an error if not set

Examples:

<!-- prettier-ignore-start -->
```yaml
commands:
  - name: dp
    cmd: docker ps
    kind: alias
    exclude: {{ notInPath "docker" }}

  - name: setup
    kind: raw
    cmd: |
      {{- if (join (env "HOME") ".rc" | exists) }}
      source {{ join (env "HOME") ".rc" }}
      {{ end }}
```
<!-- prettier-ignore-end -->

Since the entire file is rendered, templates may be used anywhere.

All paths are rendered and returned with forward slashes (`/`), even on Windows.

Examples of various use-cases can be found at [dotfiles](https://github.com/idelchi/dotfiles/tree/main/dotgen).

## Usage

```sh
dotgen [options] [patterns...]
```

- `--shell` - Target shell (default: basename of `SHELL` environment variable)
- `-f, --values` - Additional YAML variable files
- `--set` - Additional `KEY=VALUE` variables, only string values supported
- `--verbose` - Increase verbosity in rendered output
- `--debug` - Show all variables and rendered templates without processing
- `-I, --instrument` - Add instrumentation to rendered output to time commands
- `--hash` - Compute hash of all included files
- `--dry` - Show a list of files that would be processed without executing
- `-v, --version` - Show version

The positional arguments are patterns supporting globbing (`**`), with the following special cases:

- when none are provided, defaults to `**/*.dotgen`
- `.` expands to `**/*.dotgen`
- a trailing `/` expands to `**/*.dotgen` in that directory
- if a directory is provided, it expands to `**/*.dotgen` in that directory

## Use cases

**Unified dotfiles across machines**
Define once, render differently per OS/shell.

**Conditional tool integration**
Only set up starship/zoxide/fzf if they're installed. No errors on minimal systems.

**DRY shell config**
Use variables and templates instead of hardcoded paths. Change `BIN_DIR` once, update everywhere.

**Cached generation**
Generate once on login, source the cached output. Fast shell startup without losing flexibility.
Use `--hash` to further optimize caching by only regenerating on config changes.

## Demo

![Demo](assets/gifs/dotgen.gif)
