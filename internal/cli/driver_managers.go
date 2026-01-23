package cli

import "github.com/spf13/cobra"

var driverManagersCmd = &cobra.Command{
	Use:     "driver-managers",
	Aliases: []string{"driver-manager"},
	Short:   "Browse driver managers",
	Long: `Browse driver managers.

Driver managers link a manager membership to a managed membership within a trucker.

Commands:
  list    List driver managers with filtering and pagination
  show    Show full details of a driver manager`,
	Example: `  # List driver managers
  xbe view driver-managers list

  # Filter by trucker
  xbe view driver-managers list --trucker 123

  # Filter by manager membership
  xbe view driver-managers list --manager-membership 456

  # Filter by managed membership
  xbe view driver-managers list --managed-membership 789

  # Show a driver manager
  xbe view driver-managers show 321`,
}

func init() {
	viewCmd.AddCommand(driverManagersCmd)
}
