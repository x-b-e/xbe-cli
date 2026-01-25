package cli

import "github.com/spf13/cobra"

var doKeepTruckinUsersCmd = &cobra.Command{
	Use:     "keep-truckin-users",
	Aliases: []string{"keep-truckin-user"},
	Short:   "Manage KeepTruckin users",
	Long: `Manage KeepTruckin users.

Commands:
  update  Update a KeepTruckin user assignment`,
	Example: `  # Update a KeepTruckin user assignment
  xbe do keep-truckin-users update 123 --user 456`,
}

func init() {
	doCmd.AddCommand(doKeepTruckinUsersCmd)
}
