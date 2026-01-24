package cli

import "github.com/spf13/cobra"

var doBrokerCustomersCmd = &cobra.Command{
	Use:   "broker-customers",
	Short: "Manage broker customers",
	Long: `Manage broker-customer trading partner relationships.

Commands:
  create    Create a broker-customer relationship
  update    Update a broker-customer relationship
  delete    Delete a broker-customer relationship`,
	Example: `  # Create a broker-customer relationship
  xbe do broker-customers create --broker 123 --customer 456

  # Update the external accounting ID
  xbe do broker-customers update 789 --external-accounting-broker-customer-id "ACCT-42"

  # Delete a broker-customer relationship (requires --confirm)
  xbe do broker-customers delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doBrokerCustomersCmd)
}
