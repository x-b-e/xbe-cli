package cli

import "github.com/spf13/cobra"

var customerRetainersCmd = &cobra.Command{
	Use:     "customer-retainers",
	Aliases: []string{"customer-retainer"},
	Short:   "Browse customer retainers",
	Long: `Browse customer retainers.

Customer retainers track retainer agreements between a customer (buyer) and a
broker (seller).

Commands:
  list    List customer retainers with filtering
  show    Show customer retainer details`,
	Example: `  # List customer retainers
  xbe view customer-retainers list

  # Show a customer retainer
  xbe view customer-retainers show 123`,
}

func init() {
	viewCmd.AddCommand(customerRetainersCmd)
}
