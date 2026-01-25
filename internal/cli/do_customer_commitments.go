package cli

import "github.com/spf13/cobra"

var doCustomerCommitmentsCmd = &cobra.Command{
	Use:     "customer-commitments",
	Aliases: []string{"customer-commitment"},
	Short:   "Manage customer commitments",
	Long:    "Commands for creating, updating, and deleting customer commitments.",
}

func init() {
	doCmd.AddCommand(doCustomerCommitmentsCmd)
}
