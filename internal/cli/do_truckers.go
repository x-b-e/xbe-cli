package cli

import "github.com/spf13/cobra"

var doTruckersCmd = &cobra.Command{
	Use:   "truckers",
	Short: "Manage truckers",
	Long: `Manage truckers on the XBE platform.

Commands:
  create    Create a new trucker
  update    Update an existing trucker
  delete    Delete a trucker`,
	Example: `  # Create a trucker
  xbe do truckers create --name "ABC Trucking" --broker 123

  # Update a trucker
  xbe do truckers update 456 --name "New Name"

  # Delete a trucker (requires --confirm)
  xbe do truckers delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doTruckersCmd)
}
