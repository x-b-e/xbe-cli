package cli

import "github.com/spf13/cobra"

var hosDayRegulationSetsCmd = &cobra.Command{
	Use:     "hos-day-regulation-sets",
	Aliases: []string{"hos-day-regulation-set"},
	Short:   "View HOS day regulation sets",
	Long: `Commands for viewing HOS day regulation sets.

HOS day regulation sets capture the regulation set code and availability
calculations applied to a driver's HOS day.`,
	Example: `  # List HOS day regulation sets
  xbe view hos-day-regulation-sets list --limit 25

  # Show a HOS day regulation set
  xbe view hos-day-regulation-sets show 123`,
}

func init() {
	viewCmd.AddCommand(hosDayRegulationSetsCmd)
}
