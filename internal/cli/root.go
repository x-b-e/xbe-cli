package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/version"
)

var rootCmd = &cobra.Command{
	Use:           "xbe",
	Short:         "XBE CLI",
	Long:          "XBE CLI (skeleton). No functionality yet.",
	SilenceUsage:  true,
	SilenceErrors: true,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the CLI version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), version.String())
		return nil
	},
}
