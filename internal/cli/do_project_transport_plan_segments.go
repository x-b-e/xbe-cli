package cli

import "github.com/spf13/cobra"

var doProjectTransportPlanSegmentsCmd = &cobra.Command{
	Use:     "project-transport-plan-segments",
	Aliases: []string{"project-transport-plan-segment"},
	Short:   "Manage project transport plan segments",
	Long: `Manage project transport plan segments.

Segments connect origin and destination stops within a project transport plan.
You can create segments, update distance or assignments, and delete segments.
External TMS order/movement numbers are only allowed for transport-only projects.

Commands:
  create   Create a project transport plan segment
  update   Update a project transport plan segment
  delete   Delete a project transport plan segment`,
	Example: `  # Create a segment
  xbe do project-transport-plan-segments create \
    --project-transport-plan 123 \
    --origin 456 \
    --destination 789

  # Update miles
  xbe do project-transport-plan-segments update 101 --miles 12.5

  # Delete a segment
  xbe do project-transport-plan-segments delete 101 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectTransportPlanSegmentsCmd)
}
