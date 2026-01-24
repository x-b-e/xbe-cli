package cli

import "github.com/spf13/cobra"

var customerTruckersCmd = &cobra.Command{
	Use:     "customer-truckers",
	Aliases: []string{"customer-trucker"},
	Short:   "Browse customer trucker links",
	Long: `Browse customer trucker links on the XBE platform.

Customer truckers link customers to approved truckers for a broker.

Commands:
  list    List customer truckers with filtering
  show    Show customer trucker details`,
	Example: `  # List customer truckers
  xbe view customer-truckers list

  # Show a customer trucker link
  xbe view customer-truckers show 123`,
}

func init() {
	viewCmd.AddCommand(customerTruckersCmd)
}
