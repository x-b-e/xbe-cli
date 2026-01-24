package cli

import "github.com/spf13/cobra"

var doApiTokensCmd = &cobra.Command{
	Use:     "api-tokens",
	Aliases: []string{"api-token"},
	Short:   "Manage API tokens",
	Long:    "Commands for creating and updating API tokens.",
}

func init() {
	doCmd.AddCommand(doApiTokensCmd)
}
