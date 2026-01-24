package cli

import "github.com/spf13/cobra"

var projectTruckersCmd = &cobra.Command{
	Use:   "project-truckers",
	Short: "Browse project truckers",
	Long: `Browse project truckers on the XBE platform.

Project truckers link a project to a trucker and optionally exclude the trucker
from time card payroll certification requirements.

Commands:
  list    List project truckers with filtering
  show    Show project trucker details`,
	Example: `  # List project truckers
  xbe view project-truckers list

  # Filter by project
  xbe view project-truckers list --project 123

  # Filter by trucker
  xbe view project-truckers list --trucker 456

  # Filter by exclusion flag
  xbe view project-truckers list --is-excluded-from-time-card-payroll-certification-requirements true

  # Show a specific project trucker
  xbe view project-truckers show 789`,
}

func init() {
	viewCmd.AddCommand(projectTruckersCmd)
}
