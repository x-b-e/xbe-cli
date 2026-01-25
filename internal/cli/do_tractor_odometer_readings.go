package cli

import "github.com/spf13/cobra"

var doTractorOdometerReadingsCmd = &cobra.Command{
	Use:     "tractor-odometer-readings",
	Aliases: []string{"tractor-odometer-reading"},
	Short:   "Manage tractor odometer readings",
	Long:    "Commands for creating, updating, and deleting tractor odometer readings.",
}

func init() {
	doCmd.AddCommand(doTractorOdometerReadingsCmd)
}
