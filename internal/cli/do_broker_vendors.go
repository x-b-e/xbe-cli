package cli

import "github.com/spf13/cobra"

var doBrokerVendorsCmd = &cobra.Command{
	Use:   "broker-vendors",
	Short: "Manage broker vendors",
	Long: `Manage broker-vendor trading partner relationships.

Commands:
  create    Create a broker-vendor relationship
  update    Update a broker-vendor relationship
  delete    Delete a broker-vendor relationship`,
	Example: `  # Create a broker-vendor relationship
  xbe do broker-vendors create --broker 123 --vendor "Trucker|456"

  # Update the external accounting ID
  xbe do broker-vendors update 789 --external-accounting-broker-vendor-id "ACCT-42"

  # Delete a broker-vendor relationship (requires --confirm)
  xbe do broker-vendors delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doBrokerVendorsCmd)
}
