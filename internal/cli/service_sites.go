package cli

import "github.com/spf13/cobra"

var serviceSitesCmd = &cobra.Command{
	Use:   "service-sites",
	Short: "View service sites",
	Long: `View service sites on the XBE platform.

Service sites are locations used for service work orders and scheduling.

Commands:
  list    List service sites with filtering
  show    Show service site details`,
	Example: `  # List service sites
  xbe view service-sites list

  # Filter by broker
  xbe view service-sites list --broker 123

  # View service site details
  xbe view service-sites show 456`,
}

func init() {
	viewCmd.AddCommand(serviceSitesCmd)
}
