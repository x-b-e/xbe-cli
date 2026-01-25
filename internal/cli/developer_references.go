package cli

import "github.com/spf13/cobra"

var developerReferencesCmd = &cobra.Command{
	Use:     "developer-references",
	Aliases: []string{"developer-reference"},
	Short:   "View developer references",
	Long:    "Commands for viewing developer references.",
}

func init() {
	viewCmd.AddCommand(developerReferencesCmd)
}
