package cli

import "github.com/spf13/cobra"

var projectPhaseCostItemsCmd = &cobra.Command{
	Use:     "project-phase-cost-items",
	Aliases: []string{"project-phase-cost-item"},
	Short:   "View project phase cost items",
	Long:    "Commands for viewing project phase cost items.",
}

func init() {
	viewCmd.AddCommand(projectPhaseCostItemsCmd)
}
