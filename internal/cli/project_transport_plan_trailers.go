package cli

import "github.com/spf13/cobra"

var projectTransportPlanTrailersCmd = &cobra.Command{
	Use:     "project-transport-plan-trailers",
	Aliases: []string{"project-transport-plan-trailer"},
	Short:   "Browse project transport plan trailers",
	Long: `Browse project transport plan trailers.

Project transport plan trailers assign trailers across a segment range in a plan,
tracking assignment status and cached timing windows.

Commands:
  list    List trailer assignments with filtering and pagination
  show    Show full details of a trailer assignment`,
	Example: `  # List trailer assignments
  xbe view project-transport-plan-trailers list

  # Filter by project transport plan
  xbe view project-transport-plan-trailers list --project-transport-plan 123

  # Show trailer assignment details
  xbe view project-transport-plan-trailers show 456`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanTrailersCmd)
}
