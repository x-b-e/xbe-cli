package cli

import "github.com/spf13/cobra"

var doProjectTransportPlanTrailersCmd = &cobra.Command{
	Use:     "project-transport-plan-trailers",
	Aliases: []string{"project-transport-plan-trailer"},
	Short:   "Manage project transport plan trailers",
	Long: `Manage project transport plan trailers.

Trailer assignments bind trailers to a range of plan segments, tracking
assignment status and timing windows. Status defaults to editing unless set to
active (which requires a trailer).

Commands:
  create   Create a project transport plan trailer assignment
  update   Update a project transport plan trailer assignment
  delete   Delete a project transport plan trailer assignment`,
	Example: `  # Create a trailer assignment
  xbe do project-transport-plan-trailers create \
    --project-transport-plan 123 \
    --segment-start 456 \
    --segment-end 789

  # Update status
  xbe do project-transport-plan-trailers update 101 --status active --trailer 555

  # Delete a trailer assignment
  xbe do project-transport-plan-trailers delete 101 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectTransportPlanTrailersCmd)
}
