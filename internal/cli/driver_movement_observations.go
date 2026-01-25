package cli

import "github.com/spf13/cobra"

var driverMovementObservationsCmd = &cobra.Command{
	Use:     "driver-movement-observations",
	Aliases: []string{"driver-movement-observation"},
	Short:   "Browse driver movement observations",
	Long: `Browse driver movement observations.

Driver movement observations summarize driver movement cycles for a job
production plan.

Commands:
  list    List observations with filtering and pagination
  show    Show full observation details`,
	Example: `  # List driver movement observations
  xbe view driver-movement-observations list

  # Filter by job production plan
  xbe view driver-movement-observations list --plan 123

  # Show observation details
  xbe view driver-movement-observations show 456`,
}

func init() {
	viewCmd.AddCommand(driverMovementObservationsCmd)
}
