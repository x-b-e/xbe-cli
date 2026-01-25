package cli

import "github.com/spf13/cobra"

var lineupsCmd = &cobra.Command{
	Use:   "lineups",
	Short: "Browse lineups",
	Long: `Browse lineups on the XBE platform.

Lineups define scheduling windows for a customer. They are identified by a
start time range and optional name, and can group job production plans and
job schedule shifts.

Commands:
  list    List lineups
  show    Show lineup details`,
	Example: `  # List lineups
  xbe view lineups list

  # Filter by customer
  xbe view lineups list --customer 123

  # Show a lineup
  xbe view lineups show 456`,
}

func init() {
	viewCmd.AddCommand(lineupsCmd)
}
