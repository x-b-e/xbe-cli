package cli

import "github.com/spf13/cobra"

var shiftSetTimeCardConstraintsCmd = &cobra.Command{
	Use:     "shift-set-time-card-constraints",
	Aliases: []string{"shift-set-time-card-constraint"},
	Short:   "View shift set time card constraints",
	Long:    "Commands for viewing shift set time card constraints.",
}

func init() {
	viewCmd.AddCommand(shiftSetTimeCardConstraintsCmd)
}
