package cli

import "github.com/spf13/cobra"

var projectTransportPlanSegmentsCmd = &cobra.Command{
	Use:     "project-transport-plan-segments",
	Aliases: []string{"project-transport-plan-segment"},
	Short:   "Browse project transport plan segments",
	Long: `Browse project transport plan segments.

Segments connect origin and destination stops within a project transport plan,
tracking sequence position, distance, and assignments.

Commands:
  list    List segments with filtering and pagination
  show    Show full details of a segment`,
	Example: `  # List segments
  xbe view project-transport-plan-segments list

  # Filter by project transport plan
  xbe view project-transport-plan-segments list --project-transport-plan 123

  # Show segment details
  xbe view project-transport-plan-segments show 456`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanSegmentsCmd)
}
