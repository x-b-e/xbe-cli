package cli

import "github.com/spf13/cobra"

var integrationConfigsCmd = &cobra.Command{
	Use:     "integration-configs",
	Aliases: []string{"integration-config"},
	Short:   "Browse integration configs",
	Long:    "Commands for viewing integration configs.",
}

func init() {
	viewCmd.AddCommand(integrationConfigsCmd)
}
