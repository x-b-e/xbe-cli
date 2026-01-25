package cli

import "github.com/spf13/cobra"

var doSafetyIncidentsCmd = &cobra.Command{
	Use:     "safety-incidents",
	Aliases: []string{"safety-incident"},
	Short:   "Manage safety incidents",
	Long: `Commands for creating, updating, and deleting safety incidents.

Safety incidents are a specialized incident type used to track safety-related
issues, near misses, and overloading events.`,
}

func init() {
	doCmd.AddCommand(doSafetyIncidentsCmd)
}
