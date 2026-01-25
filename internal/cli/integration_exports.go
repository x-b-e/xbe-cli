package cli

import "github.com/spf13/cobra"

var integrationExportsCmd = &cobra.Command{
	Use:     "integration-exports",
	Aliases: []string{"integration-export"},
	Short:   "View integration exports",
	Long:    "Commands for viewing integration exports.",
}

func init() {
	viewCmd.AddCommand(integrationExportsCmd)
}
