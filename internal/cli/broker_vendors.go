package cli

import "github.com/spf13/cobra"

var brokerVendorsCmd = &cobra.Command{
	Use:   "broker-vendors",
	Short: "Browse broker-vendor relationships",
	Long: `Browse broker-vendor relationships.

Broker vendors represent trading partner links between a broker and a vendor
(trucker or material site). Use these commands to list, inspect, and manage
broker-vendor records.

Commands:
  list    List broker vendors with filtering and pagination
  show    Show a broker vendor by ID`,
	Example: `  # List broker-vendor relationships
  xbe view broker-vendors list

  # Filter by broker
  xbe view broker-vendors list --broker 123

  # Filter by vendor via partner filter
  xbe view broker-vendors list --partner "Trucker|456"

  # Show a broker-vendor relationship
  xbe view broker-vendors show 789`,
}

func init() {
	viewCmd.AddCommand(brokerVendorsCmd)
}
