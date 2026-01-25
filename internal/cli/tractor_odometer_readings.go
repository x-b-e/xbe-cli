package cli

import "github.com/spf13/cobra"

var tractorOdometerReadingsCmd = &cobra.Command{
	Use:     "tractor-odometer-readings",
	Aliases: []string{"tractor-odometer-reading"},
	Short:   "View tractor odometer readings",
	Long:    "Commands for viewing tractor odometer readings.",
}

func init() {
	viewCmd.AddCommand(tractorOdometerReadingsCmd)
}
