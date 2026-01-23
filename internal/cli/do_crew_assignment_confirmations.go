package cli

import "github.com/spf13/cobra"

var doCrewAssignmentConfirmationsCmd = &cobra.Command{
	Use:     "crew-assignment-confirmations",
	Aliases: []string{"crew-assignment-confirmation"},
	Short:   "Manage crew assignment confirmations",
	Long: `Commands for creating and updating crew assignment confirmations.

Note: Crew assignment confirmations cannot be deleted via the API.`,
}

func init() {
	doCmd.AddCommand(doCrewAssignmentConfirmationsCmd)
}
