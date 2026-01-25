package cli

import "github.com/spf13/cobra"

var doDeveloperReferencesCmd = &cobra.Command{
	Use:     "developer-references",
	Aliases: []string{"developer-reference"},
	Short:   "Manage developer references",
	Long:    "Commands for creating, updating, and deleting developer references.",
}

func init() {
	doCmd.AddCommand(doDeveloperReferencesCmd)
}
