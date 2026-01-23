package cli

import "github.com/spf13/cobra"

var transportReferencesCmd = &cobra.Command{
	Use:     "transport-references",
	Aliases: []string{"transport-reference"},
	Short:   "View transport references",
	Long:    "Commands for viewing transport references.",
}

func init() {
	viewCmd.AddCommand(transportReferencesCmd)
}
