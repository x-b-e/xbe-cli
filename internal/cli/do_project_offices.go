package cli

import "github.com/spf13/cobra"

var doProjectOfficesCmd = &cobra.Command{
	Use:   "project-offices",
	Short: "Manage project offices",
	Long: `Create, update, and delete project offices.

Project offices are organizational units within a broker for grouping projects.

Commands:
  create  Create a new project office
  update  Update an existing project office
  delete  Delete a project office`,
	Example: `  # Create a project office
  xbe do project-offices create --name "Chicago Office" --abbreviation "CHI" --broker 123

  # Update a project office
  xbe do project-offices update 456 --name "Updated Name"

  # Delete a project office
  xbe do project-offices delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectOfficesCmd)
}
