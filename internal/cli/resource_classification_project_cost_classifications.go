package cli

import "github.com/spf13/cobra"

var resourceClassificationProjectCostClassificationsCmd = &cobra.Command{
	Use:     "resource-classification-project-cost-classifications",
	Aliases: []string{"resource-classification-project-cost-classification"},
	Short:   "View resource classification project cost classifications",
	Long: `View resource classification project cost classifications.

Resource classification project cost classifications link labor or equipment
classifications to project cost classifications for a broker.

Commands:
  list    List resource classification project cost classifications
  show    Show resource classification project cost classification details`,
	Example: `  # List resource classification project cost classifications
  xbe view resource-classification-project-cost-classifications list

  # Show a resource classification project cost classification
  xbe view resource-classification-project-cost-classifications show 123

  # Output as JSON
  xbe view resource-classification-project-cost-classifications list --json`,
}

func init() {
	viewCmd.AddCommand(resourceClassificationProjectCostClassificationsCmd)
}
