package cli

import "github.com/spf13/cobra"

var doTransportReferencesCmd = &cobra.Command{
	Use:     "transport-references",
	Aliases: []string{"transport-reference"},
	Short:   "Manage transport references",
	Long:    "Commands for creating, updating, and deleting transport references.",
}

func init() {
	doCmd.AddCommand(doTransportReferencesCmd)
}
