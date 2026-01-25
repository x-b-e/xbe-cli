package cli

import "github.com/spf13/cobra"

var doObjectivesCmd = &cobra.Command{
	Use:   "objectives",
	Short: "Manage objectives",
	Long: `Manage objectives on the XBE platform.

Commands:
  create    Create an objective
  update    Update an objective
  delete    Delete an objective`,
	Example: `  # Create an objective
  xbe do objectives create --name "Improve On-Time Delivery" --organization "Broker|123"

  # Update an objective
  xbe do objectives update 456 --name "Updated Objective"

  # Delete an objective (requires --confirm)
  xbe do objectives delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doObjectivesCmd)
}
