package cli

import "github.com/spf13/cobra"

var doDriverDayShortfallAllocationsCmd = &cobra.Command{
	Use:     "driver-day-shortfall-allocations",
	Aliases: []string{"driver-day-shortfall-allocation"},
	Short:   "Allocate driver day shortfall quantities",
	Long:    "Commands for allocating driver day shortfall quantities across time cards.",
}

func init() {
	doCmd.AddCommand(doDriverDayShortfallAllocationsCmd)
}
