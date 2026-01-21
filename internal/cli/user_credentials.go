package cli

import "github.com/spf13/cobra"

var userCredentialsCmd = &cobra.Command{
	Use:     "user-credentials",
	Aliases: []string{"user-credential"},
	Short:   "View user credentials",
	Long:    "Commands for viewing user credentials.",
}

func init() {
	viewCmd.AddCommand(userCredentialsCmd)
}
