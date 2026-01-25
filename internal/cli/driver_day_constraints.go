package cli

import "github.com/spf13/cobra"

var driverDayConstraintsCmd = &cobra.Command{
	Use:     "driver-day-constraints",
	Aliases: []string{"driver-day-constraint"},
	Short:   "Browse driver day constraints",
	Long: `Browse driver day constraints.

Driver day constraints associate driver days with shift set time card constraints.

Commands:
  list    List driver day constraints with filtering and pagination
  show    Show full details of a driver day constraint`,
	Example: `  # List driver day constraints
  xbe view driver-day-constraints list

  # Filter by driver day
  xbe view driver-day-constraints list --driver-day 123

  # Show a driver day constraint
  xbe view driver-day-constraints show 456`,
}

func init() {
	viewCmd.AddCommand(driverDayConstraintsCmd)
}
