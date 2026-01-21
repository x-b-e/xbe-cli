package cli

import "github.com/spf13/cobra"

var projectCategoriesCmd = &cobra.Command{
	Use:   "project-categories",
	Short: "View project categories",
	Long: `View project categories on the XBE platform.

Project categories are used to classify and group projects by type of work
(e.g., paving, concrete, earthwork). They help organize reporting and filtering.

Commands:
  list    List project categories`,
	Example: `  # List project categories
  xbe view project-categories list

  # Search by name
  xbe view project-categories list --name "Paving"`,
}

func init() {
	viewCmd.AddCommand(projectCategoriesCmd)
}
