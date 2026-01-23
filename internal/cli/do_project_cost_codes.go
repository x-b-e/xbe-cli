package cli

import "github.com/spf13/cobra"

var doProjectCostCodesCmd = &cobra.Command{
	Use:     "project-cost-codes",
	Aliases: []string{"project-cost-code"},
	Short:   "Manage project cost codes",
	Long:    "Commands for creating, updating, and deleting project cost codes.",
}

func init() {
	doCmd.AddCommand(doProjectCostCodesCmd)
}
