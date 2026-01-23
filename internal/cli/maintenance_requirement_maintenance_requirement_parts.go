package cli

import "github.com/spf13/cobra"

var maintenanceRequirementMaintenanceRequirementPartsCmd = &cobra.Command{
	Use:     "maintenance-requirement-maintenance-requirement-parts",
	Aliases: []string{"maintenance-requirement-maintenance-requirement-part"},
	Short:   "View maintenance requirement parts",
	Long: `View maintenance requirement parts.

Maintenance requirement parts link maintenance requirements to required parts.

Commands:
  list    List maintenance requirement parts
  show    Show maintenance requirement part details`,
	Example: `  # List maintenance requirement parts
  xbe view maintenance-requirement-maintenance-requirement-parts list

  # Show a specific maintenance requirement part link
  xbe view maintenance-requirement-maintenance-requirement-parts show 123

  # JSON output
  xbe view maintenance-requirement-maintenance-requirement-parts list --json`,
}

func init() {
	viewCmd.AddCommand(maintenanceRequirementMaintenanceRequirementPartsCmd)
}
