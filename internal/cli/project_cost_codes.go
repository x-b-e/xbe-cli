package cli

import "github.com/spf13/cobra"

var projectCostCodesCmd = &cobra.Command{
	Use:     "project-cost-codes",
	Aliases: []string{"project-cost-code"},
	Short:   "View project cost codes",
	Long:    "Commands for viewing project cost codes.",
}

func init() {
	viewCmd.AddCommand(projectCostCodesCmd)
}
