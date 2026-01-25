package cli

import "github.com/spf13/cobra"

var projectMaterialTypeQualityControlRequirementsCmd = &cobra.Command{
	Use:   "project-material-type-quality-control-requirements",
	Short: "View project material type quality control requirements",
	Long: `View project material type quality control requirements on the XBE platform.

Project material type quality control requirements specify which quality control
classifications are required for a project material type.

Commands:
  list    List requirements with filtering
  show    Show requirement details`,
	Example: `  # List requirements
  xbe view project-material-type-quality-control-requirements list

  # Show a requirement
  xbe view project-material-type-quality-control-requirements show 123

  # Output as JSON
  xbe view project-material-type-quality-control-requirements list --json`,
}

func init() {
	viewCmd.AddCommand(projectMaterialTypeQualityControlRequirementsCmd)
}
