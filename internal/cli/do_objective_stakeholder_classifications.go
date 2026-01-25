package cli

import "github.com/spf13/cobra"

var doObjectiveStakeholderClassificationsCmd = &cobra.Command{
	Use:     "objective-stakeholder-classifications",
	Aliases: []string{"objective-stakeholder-classification"},
	Short:   "Manage objective stakeholder classifications",
	Long:    "Create, update, and delete objective stakeholder classifications.",
}

func init() {
	doCmd.AddCommand(doObjectiveStakeholderClassificationsCmd)
}
