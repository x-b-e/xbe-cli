package cli

import "github.com/spf13/cobra"

var resourceUnavailabilitiesCmd = &cobra.Command{
	Use:     "resource-unavailabilities",
	Aliases: []string{"resource-unavailability"},
	Short:   "View resource unavailabilities",
	Long:    "Commands for viewing resource unavailabilities.",
}

func init() {
	viewCmd.AddCommand(resourceUnavailabilitiesCmd)
}
