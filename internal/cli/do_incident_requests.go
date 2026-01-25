package cli

import "github.com/spf13/cobra"

var doIncidentRequestsCmd = &cobra.Command{
	Use:     "incident-requests",
	Aliases: []string{"incident-request"},
	Short:   "Manage incident requests",
	Long:    "Commands for creating, updating, and deleting incident requests.",
}

func init() {
	doCmd.AddCommand(doIncidentRequestsCmd)
}
