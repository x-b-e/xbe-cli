package cli

import "github.com/spf13/cobra"

var jobSiteTimesCmd = &cobra.Command{
	Use:   "job-site-times",
	Short: "View job site times",
	Long: `View job site times for job production plans.

Job site times track how long a user spent at a job site for a job production plan.

Commands:
  list    List job site times with filtering
  show    Show job site time details`,
	Example: `  # List job site times
  xbe view job-site-times list

  # Filter by job production plan
  xbe view job-site-times list --job-production-plan 123

  # Show a job site time
  xbe view job-site-times show 456`,
}

func init() {
	viewCmd.AddCommand(jobSiteTimesCmd)
}
