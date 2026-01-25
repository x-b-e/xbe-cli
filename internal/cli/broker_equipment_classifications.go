package cli

import "github.com/spf13/cobra"

var brokerEquipmentClassificationsCmd = &cobra.Command{
	Use:   "broker-equipment-classifications",
	Short: "View broker equipment classifications",
	Long: `View broker equipment classifications on the XBE platform.

Broker equipment classifications link brokers to equipment classifications
they can use. Equipment classifications must be non-root (have a parent).

Commands:
  list    List broker equipment classifications
  show    Show broker equipment classification details`,
	Example: `  # List broker equipment classifications
  xbe view broker-equipment-classifications list

  # Show a broker equipment classification
  xbe view broker-equipment-classifications show 123`,
}

func init() {
	viewCmd.AddCommand(brokerEquipmentClassificationsCmd)
}
