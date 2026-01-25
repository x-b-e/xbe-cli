package cli

import "github.com/spf13/cobra"

func newProjectRevenueClassificationsShowCmd() *cobra.Command {
	return newGenericShowCmd("project-revenue-classifications")
}

func init() {
	projectRevenueClassificationsCmd.AddCommand(newProjectRevenueClassificationsShowCmd())
}
