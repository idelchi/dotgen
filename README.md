# aliaser

A tool to manage and render shell aliases and functions with variable substitution.

---

[![GitHub release](https://img.shields.io/github/v/release/idelchi/aliaser)](https://github.com/idelchi/aliaser/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/idelchi/aliaser.svg)](https://pkg.go.dev/github.com/idelchi/aliaser)
[![Go Report Card](https://goreportcard.com/badge/github.com/idelchi/aliaser)](https://goreportcard.com/report/github.com/idelchi/aliaser)
[![Build Status](https://github.com/idelchi/aliaser/actions/workflows/github-actions.yml/badge.svg)](https://github.com/idelchi/aliaser/actions/workflows/github-actions.yml/badge.svg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`aliaser` is a command-line utility for defining, templating, and exporting aliases, functions, environment variables,
and custom variables into your shell.

- Write aliases and functions in a YAML configuration
- Apply Go template substitution with environment-aware defaults
- Filter commands by OS and shell
- Export usable shell-ready scripts

## Installation

```sh
curl -sSL https://raw.githubusercontent.com/idelchi/aliaser/refs/heads/main/install.sh | sh -s -- -d ~/.local/bin
```

## Usage

```sh
# Run aliaser with a config file
$ aliaser --config ~/.config/aliaser/aliaser.yaml --shell zsh
```

```sh
# Provide additional variables inline for use in templates
$ aliaser --config aliaser.yaml ENABLE_FZF_FEATURES=true KEY=VALUE
```

```sh
# Merge extra variables from YAML value files
$ aliaser --config aliaser.yaml -v values.yml
```

### Example Configuration

```yaml
# aliaser.yaml
env:
  EDITOR: nvim

vars:
  project: ~/work/src/myproject

commands:
  - name: gs
    cmd: git status
    kind: alias

  - name: greet
    kind: function
    cmd: |
      echo "Hello from '{{ .OS }}' on '{{ .ARCHITECTURE }}'!"
      echo "Project base is ${project}"
```

### Output

```sh
# Environment variables
export EDITOR=nvim

# Variables
project=~/work/src/myproject

# Commands
alias gs='git status'
greet() {
  echo "Hello from 'windows' on 'x86_64'!"
  echo "Project base is ${project}"
}
```

## Configuration

By default, `aliaser` looks for a configuration file at:

```sh
~/.config/aliaser/aliaser.yaml
```

You can override this path with the `--config` flag or the `ALIASER_CONFIG` environment variable.

### Command Kinds

- **alias** — Simple shell alias
- **function** — Multi-line shell function

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

Commands can be marked to exclude:

<!-- prettier-ignore-start -->
```yaml
commands:
  - name: debug
    cmd: echo "debug only"
    exclude: {{ eq .HOSTNAME "something" }}
```
<!-- prettier-ignore-end -->

## Variables

Variables are automatically available from:

- OS (`OS`, `PLATFORM`, `ARCHITECTURE`)
- User (`USER`, `USERNAME`, `HOME`, `HOSTNAME`)
- Active shell (`SHELL`)

You can add more through:

- Inline args: `KEY=VALUE`
- Value files via `-v file.yml`

## Template Support

All config files are processed as Go templates, extended with [slim-sprig](https://go-task.github.io/slim-sprig) functions.

Example:

```yaml
commands:
  - name: k
    cmd: kubectl --context {{.KUBECONTEXT}}
```

## Flags

- `--config`, `-c` - Set config file path (default: `$ALIASER_CONFIG`, `~/.config/aliaser/aliaser.yaml`)
- `--shell`, `-s` - Active shell name (default: `$SHELL` or `zsh`)
- `--values`, `-v` - Merge additional YAML values to use in templates
- `--show`, `-S` - Show loaded variables and exit
- `--version`, `-V` - Show version

## Demo

![Demo](assets/gifs/aliaser.gif)
