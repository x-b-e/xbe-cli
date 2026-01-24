package cli

import "github.com/spf13/cobra"

var incidentTagIncidentsCmd = &cobra.Command{
	Use:     "incident-tag-incidents",
	Aliases: []string{"incident-tag-incident"},
	Short:   "View incident tag incident links",
	Long:    "Commands for viewing incident tag incident links.",
}

func init() {
	viewCmd.AddCommand(incidentTagIncidentsCmd)
}
