package cli

import "github.com/spf13/cobra"

var doResourceClassificationProjectCostClassificationsCmd = &cobra.Command{
	Use:     "resource-classification-project-cost-classifications",
	Aliases: []string{"resource-classification-project-cost-classification"},
	Short:   "Manage resource classification project cost classifications",
	Long:    "Commands for creating and deleting resource classification project cost classifications.",
}

func init() {
	doCmd.AddCommand(doResourceClassificationProjectCostClassificationsCmd)
}
