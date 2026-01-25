package cli

import "github.com/spf13/cobra"

var doResourceUnavailabilitiesCmd = &cobra.Command{
	Use:     "resource-unavailabilities",
	Aliases: []string{"resource-unavailability"},
	Short:   "Manage resource unavailabilities",
	Long:    "Commands for creating, updating, and deleting resource unavailabilities.",
}

func init() {
	doCmd.AddCommand(doResourceUnavailabilitiesCmd)
}
