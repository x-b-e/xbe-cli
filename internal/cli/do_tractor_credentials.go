package cli

import "github.com/spf13/cobra"

var doTractorCredentialsCmd = &cobra.Command{
	Use:     "tractor-credentials",
	Aliases: []string{"tractor-credential"},
	Short:   "Manage tractor credentials",
	Long:    "Commands for creating, updating, and deleting tractor credentials.",
}

func init() {
	doCmd.AddCommand(doTractorCredentialsCmd)
}
