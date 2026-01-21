package cli

import "github.com/spf13/cobra"

var doUserCredentialsCmd = &cobra.Command{
	Use:     "user-credentials",
	Aliases: []string{"user-credential"},
	Short:   "Manage user credentials",
	Long:    "Commands for creating, updating, and deleting user credentials.",
}

func init() {
	doCmd.AddCommand(doUserCredentialsCmd)
}
