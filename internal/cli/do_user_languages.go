package cli

import "github.com/spf13/cobra"

var doUserLanguagesCmd = &cobra.Command{
	Use:     "user-languages",
	Aliases: []string{"user-language"},
	Short:   "Manage user languages",
	Long:    "Commands for creating, updating, and deleting user languages.",
}

func init() {
	doCmd.AddCommand(doUserLanguagesCmd)
}
