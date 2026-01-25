package cli

import "github.com/spf13/cobra"

var safetyIncidentsCmd = &cobra.Command{
	Use:     "safety-incidents",
	Aliases: []string{"safety-incident"},
	Short:   "View safety incidents",
	Long: `Commands for viewing safety incidents.

Safety incidents are a specialized incident type used to track safety-related
issues, near misses, and overloading events.`,
}

func init() {
	viewCmd.AddCommand(safetyIncidentsCmd)
}
