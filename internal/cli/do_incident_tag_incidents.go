package cli

import "github.com/spf13/cobra"

var doIncidentTagIncidentsCmd = &cobra.Command{
	Use:     "incident-tag-incidents",
	Aliases: []string{"incident-tag-incident"},
	Short:   "Manage incident tag incident links",
	Long:    "Commands for creating and deleting incident tag incident links.",
}

func init() {
	doCmd.AddCommand(doIncidentTagIncidentsCmd)
}
