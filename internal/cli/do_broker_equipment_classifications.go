package cli

import "github.com/spf13/cobra"

var doBrokerEquipmentClassificationsCmd = &cobra.Command{
	Use:   "broker-equipment-classifications",
	Short: "Manage broker equipment classifications",
	Long: `Manage broker equipment classifications on the XBE platform.

Broker equipment classifications link brokers to equipment classifications
they can use. Equipment classifications must be non-root (have a parent).

Commands:
  create    Create a broker equipment classification
  update    Update a broker equipment classification
  delete    Delete a broker equipment classification`,
	Example: `  # Create a broker equipment classification
  xbe do broker-equipment-classifications create --broker 123 --equipment-classification 456

  # Update a broker equipment classification
  xbe do broker-equipment-classifications update 789 --broker 123 --equipment-classification 456

  # Delete a broker equipment classification (requires --confirm)
  xbe do broker-equipment-classifications delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doBrokerEquipmentClassificationsCmd)
}
