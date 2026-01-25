package cli

import "github.com/spf13/cobra"

var doCostCodesCmd = &cobra.Command{
	Use:   "cost-codes",
	Short: "Manage cost codes",
	Long: `Create, update, and delete cost codes.

Cost codes are used to categorize and track costs for billing and accounting
purposes. They can be associated with specific customers, truckers, or brokers.

Commands:
  create    Create a new cost code
  update    Update an existing cost code
  delete    Delete a cost code`,
}

func init() {
	doCmd.AddCommand(doCostCodesCmd)
}
