package cli

import "github.com/spf13/cobra"

var expectedTimeOfArrivalsCmd = &cobra.Command{
	Use:   "expected-time-of-arrivals",
	Short: "Browse expected time of arrival updates",
	Long: `Browse expected time of arrival updates for tender job schedule shifts.

Commands:
  list    List expected time of arrivals with filtering
  show    Show expected time of arrival details`,
	Example: `  # List expected time of arrivals
  xbe view expected-time-of-arrivals list

  # Show an expected time of arrival
  xbe view expected-time-of-arrivals show 123`,
}

func init() {
	viewCmd.AddCommand(expectedTimeOfArrivalsCmd)
}
