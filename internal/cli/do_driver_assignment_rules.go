package cli

import "github.com/spf13/cobra"

var doDriverAssignmentRulesCmd = &cobra.Command{
	Use:     "driver-assignment-rules",
	Aliases: []string{"driver-assignment-rule"},
	Short:   "Manage driver assignment rules",
	Long: `Create, update, and delete driver assignment rules.

Driver assignment rules define constraints or guidance used when assigning
drivers at specific levels.`,
}

func init() {
	doCmd.AddCommand(doDriverAssignmentRulesCmd)
}
