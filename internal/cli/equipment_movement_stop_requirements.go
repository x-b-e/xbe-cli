package cli

import "github.com/spf13/cobra"

var equipmentMovementStopRequirementsCmd = &cobra.Command{
	Use:     "equipment-movement-stop-requirements",
	Aliases: []string{"equipment-movement-stop-requirement"},
	Short:   "Browse equipment movement stop requirements",
	Long: `Browse equipment movement stop requirements.

Equipment movement stop requirements link stops to equipment movement
requirements and capture whether the stop is the origin or destination
for the requirement.

Commands:
  list    List stop requirements with filtering and pagination
  show    Show full details of a stop requirement`,
	Example: `  # List stop requirements
  xbe view equipment-movement-stop-requirements list

  # Filter by stop
  xbe view equipment-movement-stop-requirements list --stop 123

  # Filter by requirement
  xbe view equipment-movement-stop-requirements list --requirement 456

  # Filter by kind
  xbe view equipment-movement-stop-requirements list --kind origin

  # Show a stop requirement
  xbe view equipment-movement-stop-requirements show 789`,
}

func init() {
	viewCmd.AddCommand(equipmentMovementStopRequirementsCmd)
}
