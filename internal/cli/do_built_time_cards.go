package cli

import "github.com/spf13/cobra"

var doBuiltTimeCardsCmd = &cobra.Command{
	Use:     "built-time-cards",
	Aliases: []string{"built-time-card"},
	Short:   "Build time cards from shifts",
	Long: `Build time cards from broker or customer tender job schedule shifts.

Built time cards derive start/end times, quantities, and down minutes from the
shift and attempt to create or update the underlying time card.

Commands:
  create    Build a time card from a shift`,
	Example: `  # Build a time card from a broker tender shift
  xbe do built-time-cards create --broker-tender-job-schedule-shift 123`,
}

func init() {
	doCmd.AddCommand(doBuiltTimeCardsCmd)
}
