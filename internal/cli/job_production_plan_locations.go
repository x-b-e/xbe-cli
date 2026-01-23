package cli

import "github.com/spf13/cobra"

var jobProductionPlanLocationsCmd = &cobra.Command{
	Use:     "job-production-plan-locations",
	Aliases: []string{"job-production-plan-location"},
	Short:   "View job production plan locations",
	Long: `View job production plan locations.

Job production plan locations represent sites associated with a job production
plan, including job sites and other locations used for routing.

Commands:
  list    List job production plan locations
  show    Show job production plan location details`,
	Example: `  # List job production plan locations
  xbe view job-production-plan-locations list

  # Show a job production plan location
  xbe view job-production-plan-locations show 123`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanLocationsCmd)
}
