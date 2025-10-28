package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// completions generates shell completion scripts for the specified shell.
func completions(cmd *cobra.Command, shell string) error {
	shells := []string{"bash", "zsh", "fish", "powershell"}

	switch shell {
	case "bash":
		_ = cmd.GenBashCompletion(cmd.OutOrStdout())
	case "zsh":
		_ = cmd.GenZshCompletion(cmd.OutOrStdout())
	case "fish":
		_ = cmd.GenFishCompletion(cmd.OutOrStdout(), true)
	case "powershell":
		_ = cmd.GenPowerShellCompletion(cmd.OutOrStdout())
	default:
		return fmt.Errorf("unsupported shell type %q, supported shells are: %v", shell, shells)
	}

	return nil
}
