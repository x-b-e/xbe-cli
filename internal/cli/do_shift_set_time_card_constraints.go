package cli

import "github.com/spf13/cobra"

var doShiftSetTimeCardConstraintsCmd = &cobra.Command{
	Use:     "shift-set-time-card-constraints",
	Aliases: []string{"shift-set-time-card-constraint"},
	Short:   "Manage shift set time card constraints",
	Long: `Create, update, and delete shift set time card constraints.

Shift set time card constraints define minimum/equality/maximum rules for
rate agreements or tenders used in time card calculations.`,
}

func init() {
	doCmd.AddCommand(doShiftSetTimeCardConstraintsCmd)
}
