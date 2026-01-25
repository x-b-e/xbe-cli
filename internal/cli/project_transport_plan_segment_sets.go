package cli

import "github.com/spf13/cobra"

var projectTransportPlanSegmentSetsCmd = &cobra.Command{
	Use:     "project-transport-plan-segment-sets",
	Aliases: []string{"project-transport-plan-segment-set"},
	Short:   "View project transport plan segment sets",
	Long: `View project transport plan segment sets.

Segment sets group segments within a project transport plan, optionally
assigned to a trucker.

Commands:
  list  List project transport plan segment sets
  show  Show project transport plan segment set details`,
	Example: `  # List project transport plan segment sets
  xbe view project-transport-plan-segment-sets list

  # Show a project transport plan segment set
  xbe view project-transport-plan-segment-sets show 123`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanSegmentSetsCmd)
}
