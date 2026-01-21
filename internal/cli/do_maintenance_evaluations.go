package cli

import "github.com/spf13/cobra"

var doMaintenanceEvaluationsCmd = &cobra.Command{
	Use:   "evaluations",
	Short: "Manage maintenance evaluations",
	Long: `Manage maintenance evaluations.

Commands:
  trigger    Trigger a maintenance requirement rule evaluation for equipment`,
	Example: `  # Trigger evaluation for equipment
  xbe do maintenance evaluations trigger --equipment-id 123`,
}

func init() {
	doMaintenanceCmd.AddCommand(doMaintenanceEvaluationsCmd)
}
