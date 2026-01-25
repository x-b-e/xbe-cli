package cli

import "github.com/spf13/cobra"

var doUserUiToursCmd = &cobra.Command{
	Use:     "user-ui-tours",
	Aliases: []string{"user-ui-tour"},
	Short:   "Manage user UI tours",
	Long:    "Commands for creating, updating, and deleting user UI tours.",
}

func init() {
	doCmd.AddCommand(doUserUiToursCmd)
}
