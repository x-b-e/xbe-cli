package cli

import "github.com/spf13/cobra"

var equipmentClassificationsCmd = &cobra.Command{
	Use:   "equipment-classifications",
	Short: "View equipment classifications",
	Long: `View equipment classifications on the XBE platform.

Equipment classifications categorize equipment types (e.g., excavators, loaders,
cranes) with their mobilization requirements. Classifications can be organized
hierarchically with parent-child relationships.

Commands:
  list    List equipment classifications`,
	Example: `  # List equipment classifications
  xbe view equipment-classifications list

  # Filter by mobilization method
  xbe view equipment-classifications list --mobilization-method trailer`,
}

func init() {
	viewCmd.AddCommand(equipmentClassificationsCmd)
}
