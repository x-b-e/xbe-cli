package cli

import "github.com/spf13/cobra"

var doBusinessUnitsCmd = &cobra.Command{
	Use:   "business-units",
	Short: "Manage business units",
	Long: `Manage business units on the XBE platform.

Commands:
  create    Create a new business unit
  update    Update an existing business unit
  delete    Delete a business unit`,
	Example: `  # Create a business unit
  xbe do business-units create --name "Paving Division" --broker 123

  # Update a business unit's name
  xbe do business-units update 456 --name "New Name"

  # Delete a business unit (requires --confirm)
  xbe do business-units delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doBusinessUnitsCmd)
}
