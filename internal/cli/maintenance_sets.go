package cli

import "github.com/spf13/cobra"

var maintenanceSetsCmd = &cobra.Command{
	Use:   "sets",
	Short: "View maintenance requirement sets",
	Long: `View maintenance requirement sets.

Requirement sets group related maintenance requirements together for a piece
of equipment. They can represent inspections, scheduled maintenance, or other
collections of maintenance tasks.

Commands:
  list    List requirement sets with filtering
  show    View detailed set information`,
	Example: `  # List all sets
  xbe view maintenance sets list

  # Filter by equipment
  xbe view maintenance sets list --equipment-id 123

  # Filter by business unit
  xbe view maintenance sets list --business-unit-id 456

  # Filter by status
  xbe view maintenance sets list --status in_progress

  # Filter by type
  xbe view maintenance sets list --type inspection

  # View set details
  xbe view maintenance sets show 789`,
}

func init() {
	maintenanceCmd.AddCommand(maintenanceSetsCmd)
}
