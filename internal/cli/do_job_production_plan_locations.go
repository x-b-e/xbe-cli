package cli

import "github.com/spf13/cobra"

var doJobProductionPlanLocationsCmd = &cobra.Command{
	Use:   "job-production-plan-locations",
	Short: "Manage job production plan locations",
	Long: `Create, update, and delete job production plan locations.

Job production plan locations represent job sites and other locations used
for routing and production plans.

Commands:
  create    Create a job production plan location
  update    Update a job production plan location
  delete    Delete a job production plan location`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanLocationsCmd)
}
