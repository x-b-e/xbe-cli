package cli

import "github.com/spf13/cobra"

func newProjectCostCodesShowCmd() *cobra.Command {
	return newGenericShowCmd("project-cost-codes")
}

func init() {
	projectCostCodesCmd.AddCommand(newProjectCostCodesShowCmd())
}
