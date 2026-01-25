package cli

import "github.com/spf13/cobra"

var objectiveStakeholderClassificationsCmd = &cobra.Command{
	Use:   "objective-stakeholder-classifications",
	Short: "View objective stakeholder classifications",
	Long: `View objective stakeholder classifications.

Objective stakeholder classifications link objective templates to stakeholder
classifications with an interest degree between 0 and 1.

Commands:
  list  List objective stakeholder classifications
  show  Show objective stakeholder classification details`,
	Example: `  # List objective stakeholder classifications
  xbe view objective-stakeholder-classifications list

  # Filter by objective
  xbe view objective-stakeholder-classifications list --objective 123

  # Show a specific objective stakeholder classification
  xbe view objective-stakeholder-classifications show 456`,
}

func init() {
	viewCmd.AddCommand(objectiveStakeholderClassificationsCmd)
}
