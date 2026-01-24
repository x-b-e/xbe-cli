package cli

import "github.com/spf13/cobra"

var apiTokensCmd = &cobra.Command{
	Use:     "api-tokens",
	Aliases: []string{"api-token"},
	Short:   "View API tokens",
	Long:    "Commands for viewing API tokens.",
}

func init() {
	viewCmd.AddCommand(apiTokensCmd)
}
