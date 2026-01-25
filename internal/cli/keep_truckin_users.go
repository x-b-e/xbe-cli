package cli

import "github.com/spf13/cobra"

var keepTruckinUsersCmd = &cobra.Command{
	Use:     "keep-truckin-users",
	Aliases: []string{"keep-truckin-user"},
	Short:   "Browse KeepTruckin users",
	Long: `Browse KeepTruckin users.

KeepTruckin users are drivers imported from the KeepTruckin integration.

Commands:
  list    List KeepTruckin users with filtering and pagination
  show    Show KeepTruckin user details`,
	Example: `  # List KeepTruckin users
  xbe view keep-truckin-users list

  # Show a KeepTruckin user
  xbe view keep-truckin-users show 123`,
}

func init() {
	viewCmd.AddCommand(keepTruckinUsersCmd)
}
