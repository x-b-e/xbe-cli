package cli

import "github.com/spf13/cobra"

var doProjectTruckersCmd = &cobra.Command{
	Use:   "project-truckers",
	Short: "Manage project truckers",
	Long: `Manage project truckers on the XBE platform.

Project truckers link a project to a trucker and can exclude the trucker from
certain time card payroll certification requirements.

Commands:
  create    Create a project trucker
  update    Update a project trucker
  delete    Delete a project trucker`,
	Example: `  # Create a project trucker
  xbe do project-truckers create --project 123 --trucker 456

  # Update the exclusion flag
  xbe do project-truckers update 789 --is-excluded-from-time-card-payroll-certification-requirements=true

  # Delete a project trucker (requires --confirm)
  xbe do project-truckers delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectTruckersCmd)
}
