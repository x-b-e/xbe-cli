package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/version"
)

var rootCmd = &cobra.Command{
	Use:   "xbe",
	Short: "XBE CLI - Access XBE platform data and services",
	Long: `XBE CLI - Access XBE platform data and services

The XBE command-line interface provides programmatic access to the XBE platform,
enabling you to browse newsletters, manage broker data, and integrate XBE
capabilities into your workflows.

This CLI is designed for both interactive use and automation. All commands
support JSON output (--json) for easy parsing and integration with other tools.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func init() {
	initHelp(rootCmd)
	rootCmd.AddCommand(versionCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the CLI version",
	Long: `Print the CLI version.

Displays the current version of the XBE CLI. Useful for debugging,
reporting issues, or verifying you have the latest version installed.`,
	Example: `  # Show version
  xbe version`,
	Annotations: map[string]string{"group": GroupUtility},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), version.String())
		return nil
	},
}
