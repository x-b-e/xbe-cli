package cli

import "github.com/spf13/cobra"

var doBrokerCertificationTypesCmd = &cobra.Command{
	Use:   "broker-certification-types",
	Short: "Manage broker certification types",
	Long: `Manage broker certification types on the XBE platform.

Broker certification types link brokers to certification types they can
track or require.

Commands:
  create    Create a broker certification type
  update    Update a broker certification type
  delete    Delete a broker certification type`,
	Example: `  # Create a broker certification type
  xbe do broker-certification-types create --broker 123 --certification-type 456

  # Update a broker certification type
  xbe do broker-certification-types update 789 --broker 123 --certification-type 456

  # Delete a broker certification type (requires --confirm)
  xbe do broker-certification-types delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doBrokerCertificationTypesCmd)
}
