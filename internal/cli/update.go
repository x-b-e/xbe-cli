package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

type updateOutput struct {
	Command string `json:"command"`
	Tag     string `json:"tag,omitempty"`
	Method  string `json:"method"`
	URL     string `json:"url,omitempty"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Show update instructions",
	Long: `Show instructions for updating the XBE CLI to the latest version.

Displays the appropriate update command for your platform:
  - macOS/Linux: Shell script that downloads and installs the latest release
  - Windows: Manual download instructions

You can optionally pin to a specific version using the --tag flag.`,
	Example: `  # Show update command for latest version
  xbe update

  # Pin to a specific version
  xbe update --tag v0.1.0

  # Get update info as JSON (for automation)
  xbe update --json`,
	Annotations: map[string]string{"group": GroupUtility},
	RunE:        runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().String("tag", "", "Pin to a specific release tag (e.g., v0.1.0)")
	updateCmd.Flags().Bool("json", false, "Output JSON")
}

func runUpdate(cmd *cobra.Command, _ []string) error {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}
	tag, err := cmd.Flags().GetString("tag")
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		if jsonOut {
			return writeJSON(cmd.OutOrStdout(), updateOutput{
				Method: "manual",
				URL:    "https://github.com/x-b-e/xbe-cli/releases/latest",
			})
		}
		fmt.Fprintln(cmd.OutOrStdout(), "On Windows, download the latest release zip and replace xbe.exe.")
		fmt.Fprintln(cmd.OutOrStdout(), "https://github.com/x-b-e/xbe-cli/releases/latest")
		return nil
	}

	scriptURL := "https://raw.githubusercontent.com/x-b-e/xbe-cli/main/scripts/install.sh"
	command := fmt.Sprintf("curl -fsSL %s | bash", scriptURL)
	if tag != "" {
		command = fmt.Sprintf("TAG=%s %s", tag, command)
	}

	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), updateOutput{
			Command: command,
			Tag:     tag,
			Method:  "script",
		})
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Run this to update:")
	fmt.Fprintln(cmd.OutOrStdout(), command)
	return nil
}
