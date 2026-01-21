package cli

import "github.com/spf13/cobra"

var doTruckScopesCmd = &cobra.Command{
	Use:   "truck-scopes",
	Short: "Manage truck scopes",
	Long: `Create, update, and delete truck scopes.

Truck scopes define geographic and equipment restrictions for trucking operations.

Commands:
  create  Create a new truck scope
  update  Update an existing truck scope
  delete  Delete a truck scope`,
	Example: `  # Create a truck scope
  xbe do truck-scopes create --authorized-state-codes "IL,IN,WI"

  # Update a truck scope
  xbe do truck-scopes update 456 --authorized-state-codes "IL,IN,WI,MI"

  # Delete a truck scope
  xbe do truck-scopes delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doTruckScopesCmd)
}
