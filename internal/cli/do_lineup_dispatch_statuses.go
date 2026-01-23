package cli

import "github.com/spf13/cobra"

var doLineupDispatchStatusesCmd = &cobra.Command{
	Use:     "lineup-dispatch-statuses",
	Aliases: []string{"lineup-dispatch-status"},
	Short:   "Check lineup dispatch offer status",
	Long: `Check lineup dispatch offer status for a broker window and date.

Commands:
  create    Compute the offered tender percentage for a lineup window`,
}

func init() {
	doCmd.AddCommand(doLineupDispatchStatusesCmd)
}
