package cli

import "github.com/spf13/cobra"

var projectTransportPlanStopInsertionsCmd = &cobra.Command{
	Use:     "project-transport-plan-stop-insertions",
	Aliases: []string{"project-transport-plan-stop-insertion"},
	Short:   "View project transport plan stop insertions",
	Long: `View project transport plan stop insertions.

Stop insertions apply insert, move, or delete operations to project transport
plan stops while tracking status, results, and error details.

Commands:
  list  List project transport plan stop insertions
  show  Show project transport plan stop insertion details`,
	Example: `  # List stop insertions
  xbe view project-transport-plan-stop-insertions list --limit 10

  # Show a stop insertion
  xbe view project-transport-plan-stop-insertions show 123`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanStopInsertionsCmd)
}
