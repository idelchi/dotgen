// Package exec provides utilities to run shell commands and capture their output.
package exec

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"
)

// Result represents the result of executing a shell command.
type Result struct {
	// Stdout is the standard output of the command.
	Stdout string
	// Stderr is the standard error output of the command.
	Stderr string
	// ExitCode is the exit code of the command.
	ExitCode int
	// Err is any error that occurred during command execution.
	Err error
}

// Run executes the given shell command snippet using the specified shell.
func Run(shell, snippet string, timeout time.Duration) *Result {
	if strings.TrimSpace(shell) == "" {
		return &Result{
			ExitCode: -1,
			Err:      errors.New("active shell is required"),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, shell, "-c", snippet)

	var outBuf, errBuf bytes.Buffer

	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()

	exitCode := -1

	if ps := cmd.ProcessState; ps != nil {
		exitCode = ps.ExitCode()
	}

	return &Result{
		Stdout:   outBuf.String(),
		Stderr:   errBuf.String(),
		ExitCode: exitCode,
		Err:      err,
	}
}
