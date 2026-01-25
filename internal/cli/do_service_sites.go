package cli

import "github.com/spf13/cobra"

var doServiceSitesCmd = &cobra.Command{
	Use:   "service-sites",
	Short: "Manage service sites",
	Long: `Create, update, and delete service sites.

Service sites are locations used for service work orders.

Commands:
  create    Create a new service site
  update    Update an existing service site
  delete    Delete a service site`,
}

func init() {
	doCmd.AddCommand(doServiceSitesCmd)
}
