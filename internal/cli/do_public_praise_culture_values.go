package cli

import "github.com/spf13/cobra"

var doPublicPraiseCultureValuesCmd = &cobra.Command{
	Use:     "public-praise-culture-values",
	Aliases: []string{"public-praise-culture-value"},
	Short:   "Manage public praise culture values",
	Long:    "Commands for creating, updating, and deleting public praise culture values.",
}

func init() {
	doCmd.AddCommand(doPublicPraiseCultureValuesCmd)
}
