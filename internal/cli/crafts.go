package cli

import "github.com/spf13/cobra"

var craftsCmd = &cobra.Command{
	Use:   "crafts",
	Short: "View crafts",
	Long: `View crafts on the XBE platform.

Crafts define trade classifications for workers (e.g., carpenter, electrician)
and are scoped to a broker organization.

Commands:
  list    List crafts`,
	Example: `  # List crafts
  xbe view crafts list

  # Filter by broker
  xbe view crafts list --broker 123

  # Output as JSON
  xbe view crafts list --json`,
}

func init() {
	viewCmd.AddCommand(craftsCmd)
}
