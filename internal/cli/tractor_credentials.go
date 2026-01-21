package cli

import "github.com/spf13/cobra"

var tractorCredentialsCmd = &cobra.Command{
	Use:     "tractor-credentials",
	Aliases: []string{"tractor-credential"},
	Short:   "View tractor credentials",
	Long:    "Commands for viewing tractor credentials.",
}

func init() {
	viewCmd.AddCommand(tractorCredentialsCmd)
}
