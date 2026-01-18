package cli

import "github.com/spf13/cobra"

var jobProductionPlansCmd = &cobra.Command{
	Use:   "job-production-plans",
	Short: "View job production plans",
	Long: `View job production plans on the XBE platform.

Job production plans are the core scheduling unit in XBE, representing a day's
work at a specific job site with material deliveries, crew assignments, and
production targets.

Commands:
  list    List job production plans with filtering
  show    Show details of a specific plan`,
	Example: `  # List plans for today
  xbe view job-production-plans list --start-on 2025-01-18

  # List plans for a date range
  xbe view job-production-plans list --start-on 2025-01-01 --end-on 2025-01-31

  # Filter by status
  xbe view job-production-plans list --start-on 2025-01-18 --status approved

  # Search by job name or number
  xbe view job-production-plans list --start-on 2025-01-18 --q "Main Street"

  # Show plan details
  xbe view job-production-plans show 12345`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlansCmd)
}
