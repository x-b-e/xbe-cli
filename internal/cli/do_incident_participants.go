package cli

import "github.com/spf13/cobra"

var doIncidentParticipantsCmd = &cobra.Command{
	Use:     "incident-participants",
	Aliases: []string{"incident-participant"},
	Short:   "Manage incident participants",
	Long:    "Commands for creating, updating, and deleting incident participants.",
}

func init() {
	doCmd.AddCommand(doIncidentParticipantsCmd)
}
