package cli

import "github.com/spf13/cobra"

var materialTypesCmd = &cobra.Command{
	Use:   "material-types",
	Short: "View material types",
	Long: `View material types on the XBE platform.

Material types define the materials that can be hauled or used in jobs (e.g.,
gravel, asphalt, concrete). They can be organized hierarchically with parent
types and sub-types, and may be associated with specific suppliers.

Commands:
  list    List material types`,
	Example: `  # List material types
  xbe view material-types list

  # Search by name
  xbe view material-types list --name "gravel"`,
}

func init() {
	viewCmd.AddCommand(materialTypesCmd)
}
