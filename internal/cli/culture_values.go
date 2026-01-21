package cli

import "github.com/spf13/cobra"

var cultureValuesCmd = &cobra.Command{
	Use:   "culture-values",
	Short: "View culture values",
	Long: `View culture values on the XBE platform.

Culture values define organizational values used for public praise and
recognition. They help reinforce company culture and values.

Commands:
  list    List culture values`,
	Example: `  # List culture values
  xbe view culture-values list

  # Filter by organization
  xbe view culture-values list --organization 123`,
}

func init() {
	viewCmd.AddCommand(cultureValuesCmd)
}
