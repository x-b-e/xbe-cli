package cli

import "github.com/spf13/cobra"

func newUsersShowCmd() *cobra.Command {
	return newGenericShowCmd("users")
}

func init() {
	usersCmd.AddCommand(newUsersShowCmd())
}
