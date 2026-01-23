package cli

import "github.com/spf13/cobra"

var doProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage projects",
	Long: `Manage projects on the XBE platform.

Commands:
  create    Create a new project
  update    Update an existing project
  delete    Delete a project`,
	Example: `  # Create a project
  xbe do projects create --name "Highway 101 Expansion" --developer 123

  # Update a project
  xbe do projects update 456 --name "Updated Name"

  # Delete a project (requires --confirm)
  xbe do projects delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectsCmd)
}
