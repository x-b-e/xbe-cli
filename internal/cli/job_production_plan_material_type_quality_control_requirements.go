package cli

import "github.com/spf13/cobra"

var jobProductionPlanMaterialTypeQualityControlRequirementsCmd = &cobra.Command{
	Use:     "job-production-plan-material-type-quality-control-requirements",
	Aliases: []string{"job-production-plan-material-type-quality-control-requirement"},
	Short:   "Browse job production plan material type quality control requirements",
	Long: `Browse job production plan material type quality control requirements on the XBE platform.

These requirements link a job production plan material type to a quality control
classification, with an optional note.

Commands:
  list    List requirements with filtering and pagination
  show    Show requirement details`,
	Example: `  # List requirements
  xbe view job-production-plan-material-type-quality-control-requirements list

  # Show a requirement
  xbe view job-production-plan-material-type-quality-control-requirements show 123

  # Output as JSON
  xbe view job-production-plan-material-type-quality-control-requirements list --json`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanMaterialTypeQualityControlRequirementsCmd)
}
