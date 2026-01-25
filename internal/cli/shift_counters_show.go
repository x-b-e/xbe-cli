package cli

import "github.com/spf13/cobra"

func newShiftCountersShowCmd() *cobra.Command {
	return newGenericShowCmd("shift-counters")
}

func init() {
	shiftCountersCmd.AddCommand(newShiftCountersShowCmd())
}
