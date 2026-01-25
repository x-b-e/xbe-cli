package cli

import "github.com/spf13/cobra"

var projectTransportPlanDriverConfirmationsCmd = &cobra.Command{
	Use:     "project-transport-plan-driver-confirmations",
	Aliases: []string{"project-transport-plan-driver-confirmation"},
	Short:   "Browse project transport plan driver confirmations",
	Long: `Browse project transport plan driver confirmations.

Driver confirmations track when a driver confirms (or rejects) a project
transport plan assignment.

Commands:
  list    List confirmations with filtering and pagination
  show    Show full details of a confirmation`,
	Example: `  # List confirmations
  xbe view project-transport-plan-driver-confirmations list

  # Filter by status
  xbe view project-transport-plan-driver-confirmations list --status pending

  # Filter by project transport plan
  xbe view project-transport-plan-driver-confirmations list --project-transport-plan 123

  # Filter by driver
  xbe view project-transport-plan-driver-confirmations list --driver 456

  # Show confirmation details
  xbe view project-transport-plan-driver-confirmations show 789`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanDriverConfirmationsCmd)
}
