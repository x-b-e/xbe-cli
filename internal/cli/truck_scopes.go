package cli

import "github.com/spf13/cobra"

var truckScopesCmd = &cobra.Command{
	Use:   "truck-scopes",
	Short: "View truck scopes",
	Long: `View truck scopes.

Truck scopes define geographic and equipment restrictions for trucking operations.

Commands:
  list  List truck scopes`,
}

func init() {
	viewCmd.AddCommand(truckScopesCmd)
}
