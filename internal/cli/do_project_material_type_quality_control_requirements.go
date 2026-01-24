package cli

import "github.com/spf13/cobra"

var doProjectMaterialTypeQualityControlRequirementsCmd = &cobra.Command{
	Use:   "project-material-type-quality-control-requirements",
	Short: "Manage project material type quality control requirements",
	Long: `Create, update, and delete project material type quality control requirements.

Project material type quality control requirements specify which quality control
classifications are required for a project material type.

Commands:
  create    Create a new requirement
  update    Update an existing requirement
  delete    Delete a requirement`,
}

func init() {
	doCmd.AddCommand(doProjectMaterialTypeQualityControlRequirementsCmd)
}
