package cli

import "github.com/spf13/cobra"

var doProjectTrailerClassificationsCmd = &cobra.Command{
	Use:     "project-trailer-classifications",
	Aliases: []string{"project-trailer-classification"},
	Short:   "Manage project trailer classifications",
	Long: `Create, update, and delete project trailer classifications.

Project trailer classifications associate trailer classifications with projects.
They can optionally link to project labor classifications.

Commands:
  create  Create a project trailer classification
  update  Update a project trailer classification
  delete  Delete a project trailer classification`,
	Example: `  # Create a project trailer classification
  xbe do project-trailer-classifications create --project 123 --trailer-classification 456

  # Update project labor classification
  xbe do project-trailer-classifications update 789 --project-labor-classification 321

  # Delete a project trailer classification
  xbe do project-trailer-classifications delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectTrailerClassificationsCmd)
}
