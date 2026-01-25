package cli

import "github.com/spf13/cobra"

var doDownMinutesEstimatesCmd = &cobra.Command{
	Use:     "down-minutes-estimates",
	Aliases: []string{"down-minutes-estimate"},
	Short:   "Estimate down minutes for a shift",
	Long: `Estimate down minutes for a tender job schedule shift.

Down minutes estimates use production incident windows to compute the estimated
and credited down minutes for a shift.

Commands:
  create    Estimate down minutes for a shift`,
	Example: `  # Estimate down minutes for a tender job schedule shift
  xbe do down-minutes-estimates create --tender-job-schedule-shift 123`,
}

func init() {
	doCmd.AddCommand(doDownMinutesEstimatesCmd)
}
