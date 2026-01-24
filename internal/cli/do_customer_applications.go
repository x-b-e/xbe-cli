package cli

import "github.com/spf13/cobra"

var doCustomerApplicationsCmd = &cobra.Command{
	Use:     "customer-applications",
	Aliases: []string{"customer-application"},
	Short:   "Manage customer applications",
	Long:    "Commands for creating, updating, and deleting customer applications.",
}

func init() {
	doCmd.AddCommand(doCustomerApplicationsCmd)
}
