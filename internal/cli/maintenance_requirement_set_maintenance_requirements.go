package cli

import "github.com/spf13/cobra"

var maintenanceRequirementSetMaintenanceRequirementsCmd = &cobra.Command{
	Use:     "maintenance-requirement-set-maintenance-requirements",
	Aliases: []string{"maintenance-requirement-set-maintenance-requirement"},
	Short:   "Browse maintenance requirement set maintenance requirements",
	Long: `Browse maintenance requirement set maintenance requirements.

These records link maintenance requirements to maintenance requirement sets.

Commands:
  list    List records with filters
  show    Show record details`,
	Example: `  # List records
  xbe view maintenance-requirement-set-maintenance-requirements list

  # Show details for a record
  xbe view maintenance-requirement-set-maintenance-requirements show 123`,
}

func init() {
	viewCmd.AddCommand(maintenanceRequirementSetMaintenanceRequirementsCmd)
}
