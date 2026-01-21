package cli

import "github.com/spf13/cobra"

var doTrailerCredentialsCmd = &cobra.Command{
	Use:     "trailer-credentials",
	Aliases: []string{"trailer-credential"},
	Short:   "Manage trailer credentials",
	Long:    "Commands for creating, updating, and deleting trailer credentials.",
}

func init() {
	doCmd.AddCommand(doTrailerCredentialsCmd)
}
