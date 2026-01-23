package cli

import "github.com/spf13/cobra"

var doDriverDayConstraintsCmd = &cobra.Command{
	Use:     "driver-day-constraints",
	Aliases: []string{"driver-day-constraint"},
	Short:   "Manage driver day constraints",
	Long: `Create, update, and delete driver day constraints.

Driver day constraints associate driver days with shift set time card constraints.`,
}

func init() {
	doCmd.AddCommand(doDriverDayConstraintsCmd)
}
