package cli

import "github.com/spf13/cobra"

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Browse and view projects",
	Long: `Browse and view projects on the XBE platform.

Projects represent construction or delivery projects that organize
job sites, materials, and logistics operations.

Commands:
  list    List projects with filtering and pagination`,
	Example: `  # List projects
  xbe view projects list

  # Search by project name
  xbe view projects list --name "Highway"

  # Filter by status
  xbe view projects list --status active

  # Get results as JSON
  xbe view projects list --json --limit 10`,
}

func init() {
	viewCmd.AddCommand(projectsCmd)
}
