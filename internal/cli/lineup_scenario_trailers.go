package cli

import "github.com/spf13/cobra"

var lineupScenarioTrailersCmd = &cobra.Command{
	Use:     "lineup-scenario-trailers",
	Aliases: []string{"lineup-scenario-trailer"},
	Short:   "View lineup scenario trailers",
	Long:    "Commands for viewing lineup scenario trailers.",
}

func init() {
	viewCmd.AddCommand(lineupScenarioTrailersCmd)
}
