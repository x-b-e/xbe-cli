package cli

import "github.com/spf13/cobra"

var jobProductionPlanTrailerClassificationsCmd = &cobra.Command{
	Use:     "job-production-plan-trailer-classifications",
	Aliases: []string{"job-production-plan-trailer-classification"},
	Short:   "View job production plan trailer classifications",
	Long: `View job production plan trailer classifications.

Job production plan trailer classifications define which trailer
classifications apply to a job production plan, including weight and
material transaction limits.

Commands:
  list    List job production plan trailer classifications
  show    Show job production plan trailer classification details`,
	Example: `  # List job production plan trailer classifications
  xbe view job-production-plan-trailer-classifications list

  # Show a job production plan trailer classification
  xbe view job-production-plan-trailer-classifications show 123`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanTrailerClassificationsCmd)
}
