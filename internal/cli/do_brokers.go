package cli

import "github.com/spf13/cobra"

var doBrokersCmd = &cobra.Command{
	Use:   "brokers",
	Short: "Manage brokers",
	Long: `Manage brokers on the XBE platform.

Commands:
  create    Create a new broker
  update    Update an existing broker
  delete    Delete a broker`,
	Example: `  # Create a broker
  xbe do brokers create --name "ABC Logistics"

  # Update a broker
  xbe do brokers update 123 --name "New Name"

  # Delete a broker (requires --confirm)
  xbe do brokers delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doBrokersCmd)
}
