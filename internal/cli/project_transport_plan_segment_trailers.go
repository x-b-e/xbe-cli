package cli

import "github.com/spf13/cobra"

var projectTransportPlanSegmentTrailersCmd = &cobra.Command{
	Use:   "project-transport-plan-segment-trailers",
	Short: "Browse project transport plan segment trailers",
	Long: `Browse project transport plan segment trailers on the XBE platform.

Project transport plan segment trailers associate trailers with transport plan segments.

Commands:
  list  List project transport plan segment trailers
  show  Show project transport plan segment trailer details`,
	Example: `  # List project transport plan segment trailers
  xbe view project-transport-plan-segment-trailers list --limit 10

  # Show a project transport plan segment trailer
  xbe view project-transport-plan-segment-trailers show 123`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanSegmentTrailersCmd)
}
