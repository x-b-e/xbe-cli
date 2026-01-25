package cli

import "github.com/spf13/cobra"

var publicPraisesCmd = &cobra.Command{
	Use:     "public-praises",
	Aliases: []string{"public-praise"},
	Short:   "View public praises",
	Long:    "Commands for viewing public praises (employee recognition).",
}

func init() {
	viewCmd.AddCommand(publicPraisesCmd)
}
