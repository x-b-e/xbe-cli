package cli

import "github.com/spf13/cobra"

var doRetainersCmd = &cobra.Command{
	Use:   "retainers",
	Short: "Manage retainers",
	Long: `Create, update, and delete retainers.

Retainers define ongoing agreements between buyers and sellers,
including expected earnings and travel limits.

Commands:
  create  Create a new retainer
  update  Update an existing retainer
  delete  Delete a retainer`,
	Example: `  # Create a broker retainer (broker -> trucker)
  xbe do retainers create --buyer Broker|123 --seller Trucker|456 --status editing

  # Create a customer retainer (customer -> broker)
  xbe do retainers create --buyer Customer|789 --seller Broker|123 --status editing

  # Update travel limits
  xbe do retainers update 456 --maximum-travel-minutes 90

  # Delete a retainer
  xbe do retainers delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doRetainersCmd)
}
