package cli

import "github.com/spf13/cobra"

var doDriverManagersCmd = &cobra.Command{
	Use:   "driver-managers",
	Short: "Manage driver managers",
	Long: `Manage driver managers.

Commands:
  create    Create a driver manager
  update    Update a driver manager
  delete    Delete a driver manager`,
	Example: `  # Create a driver manager
  xbe do driver-managers create \
    --trucker 123 \
    --manager-membership 456 \
    --managed-membership 789

  # Update a driver manager
  xbe do driver-managers update 321 --manager-membership 456

  # Delete a driver manager
  xbe do driver-managers delete 321 --confirm`,
}

func init() {
	doCmd.AddCommand(doDriverManagersCmd)
}
