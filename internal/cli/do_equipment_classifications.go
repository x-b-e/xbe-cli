package cli

import "github.com/spf13/cobra"

var doEquipmentClassificationsCmd = &cobra.Command{
	Use:   "equipment-classifications",
	Short: "Manage equipment classifications",
	Long: `Manage equipment classifications on the XBE platform.

Equipment classifications define types of equipment (e.g., paver, roller, loader)
with their mobilization requirements. Classifications can be hierarchical with
parent-child relationships.

Commands:
  create    Create a new equipment classification
  update    Update an existing equipment classification
  delete    Delete an equipment classification`,
	Example: `  # Create an equipment classification
  xbe do equipment-classifications create --name "Paver" --abbreviation "paver"

  # Create a child classification
  xbe do equipment-classifications create --name "Asphalt Paver" --abbreviation "asph-paver" --parent 123

  # Update an equipment classification
  xbe do equipment-classifications update 456 --name "Updated Name"

  # Delete an equipment classification (requires --confirm)
  xbe do equipment-classifications delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doEquipmentClassificationsCmd)
}
