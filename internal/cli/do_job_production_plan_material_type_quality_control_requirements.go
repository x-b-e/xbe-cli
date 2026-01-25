package cli

import "github.com/spf13/cobra"

var doJobProductionPlanMaterialTypeQualityControlRequirementsCmd = &cobra.Command{
	Use:     "job-production-plan-material-type-quality-control-requirements",
	Aliases: []string{"job-production-plan-material-type-quality-control-requirement"},
	Short:   "Manage job production plan material type quality control requirements",
	Long: `Manage job production plan material type quality control requirements.

Commands:
  create    Create a quality control requirement
  update    Update a quality control requirement
  delete    Delete a quality control requirement`,
	Example: `  # Create a requirement
  xbe do job-production-plan-material-type-quality-control-requirements create \
    --job-production-plan-material-type 123 \
    --quality-control-classification 456 \
    --note "Temperature check"

  # Update a requirement
  xbe do job-production-plan-material-type-quality-control-requirements update 789 --note "Updated note"

  # Delete a requirement
  xbe do job-production-plan-material-type-quality-control-requirements delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanMaterialTypeQualityControlRequirementsCmd)
}
