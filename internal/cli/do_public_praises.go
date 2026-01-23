package cli

import "github.com/spf13/cobra"

var doPublicPraisesCmd = &cobra.Command{
	Use:     "public-praises",
	Aliases: []string{"public-praise"},
	Short:   "Manage public praises",
	Long:    "Commands for creating, updating, and deleting public praises.",
}

func init() {
	doCmd.AddCommand(doPublicPraisesCmd)
}
