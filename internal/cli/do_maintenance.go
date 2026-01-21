package cli

import "github.com/spf13/cobra"

var doMaintenanceCmd = &cobra.Command{
	Use:   "maintenance",
	Short: "Manage maintenance evaluations",
	Long: `Manage maintenance evaluations on the XBE platform.

Commands:
  evaluations   Trigger maintenance requirement rule evaluations`,
	Example: `  # Trigger evaluation for equipment
  xbe do maintenance evaluations trigger --equipment-id 123`,
}

func init() {
	doCmd.AddCommand(doMaintenanceCmd)
}
