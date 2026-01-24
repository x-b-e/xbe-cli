package cli

import "github.com/spf13/cobra"

var publicPraiseCultureValuesCmd = &cobra.Command{
	Use:     "public-praise-culture-values",
	Aliases: []string{"public-praise-culture-value"},
	Short:   "View public praise culture values",
	Long: `View public praise culture values.

Public praise culture values link public praises to culture values, indicating
which values a praise recognizes.

Commands:
  list    List public praise culture values
  show    Show public praise culture value details`,
}

func init() {
	viewCmd.AddCommand(publicPraiseCultureValuesCmd)
}
