package cli

import "github.com/spf13/cobra"

func newProjectCostClassificationsShowCmd() *cobra.Command {
	return newGenericShowCmd("project-cost-classifications")
}

func init() {
	projectCostClassificationsCmd.AddCommand(newProjectCostClassificationsShowCmd())
}
