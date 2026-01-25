package cli

import "github.com/spf13/cobra"

var incidentsCmd = &cobra.Command{
	Use:     "incidents",
	Aliases: []string{"incident"},
	Short:   "View incidents",
	Long: `Commands for viewing incidents.

Incidents can be of various types: safety, production, efficiency, liability, administrative.
Create operations require using specific incident type resources (e.g., safety-incidents).`,
}

func init() {
	viewCmd.AddCommand(incidentsCmd)
}
