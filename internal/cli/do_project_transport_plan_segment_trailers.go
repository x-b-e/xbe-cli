package cli

import "github.com/spf13/cobra"

var doProjectTransportPlanSegmentTrailersCmd = &cobra.Command{
	Use:   "project-transport-plan-segment-trailers",
	Short: "Manage project transport plan segment trailers",
	Long: `Create and delete project transport plan segment trailers.

Project transport plan segment trailers associate trailers with transport plan segments.

Commands:
  create  Create a project transport plan segment trailer
  delete  Delete a project transport plan segment trailer`,
	Example: `  # Create a project transport plan segment trailer
  xbe do project-transport-plan-segment-trailers create \
    --project-transport-plan-segment 123 \
    --trailer 456

  # Delete a project transport plan segment trailer
  xbe do project-transport-plan-segment-trailers delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectTransportPlanSegmentTrailersCmd)
}
