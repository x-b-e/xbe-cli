package cli

import "github.com/spf13/cobra"

var maintenanceRequirementSetsCmd = &cobra.Command{
	Use:     "maintenance-requirement-sets",
	Aliases: []string{"maintenance-requirement-set"},
	Short:   "View maintenance requirement sets",
	Long: `View maintenance requirement sets.

Maintenance requirement sets group related maintenance requirements for equipment.

Commands:
  list    List maintenance requirement sets
  show    Show maintenance requirement set details`,
	Example: `  # List maintenance requirement sets
  xbe view maintenance-requirement-sets list

  # Show a maintenance requirement set
  xbe view maintenance-requirement-sets show 123

  # JSON output
  xbe view maintenance-requirement-sets list --json`,
}

func init() {
	viewCmd.AddCommand(maintenanceRequirementSetsCmd)
}
