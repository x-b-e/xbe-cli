package cli

import "github.com/spf13/cobra"

var doExpectedTimeOfArrivalsCmd = &cobra.Command{
	Use:   "expected-time-of-arrivals",
	Short: "Manage expected time of arrivals",
	Long: `Manage expected time of arrival updates for tender job schedule shifts.

Commands:
  create    Create an expected time of arrival
  update    Update an expected time of arrival
  delete    Delete an expected time of arrival`,
	Example: `  # Create an expected time of arrival
  xbe do expected-time-of-arrivals create --tender-job-schedule-shift 123 --expected-at 2025-01-15T12:00:00Z

  # Update an expected time of arrival
  xbe do expected-time-of-arrivals update 456 --note "Running late"

  # Delete an expected time of arrival (requires --confirm)
  xbe do expected-time-of-arrivals delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doExpectedTimeOfArrivalsCmd)
}
