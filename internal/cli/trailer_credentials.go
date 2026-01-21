package cli

import "github.com/spf13/cobra"

var trailerCredentialsCmd = &cobra.Command{
	Use:     "trailer-credentials",
	Aliases: []string{"trailer-credential"},
	Short:   "View trailer credentials",
	Long:    "Commands for viewing trailer credentials.",
}

func init() {
	viewCmd.AddCommand(trailerCredentialsCmd)
}
