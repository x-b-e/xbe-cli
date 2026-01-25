package cli

import "github.com/spf13/cobra"

var doProjectLaborClassificationsCmd = &cobra.Command{
	Use:   "project-labor-classifications",
	Short: "Manage project labor classifications",
	Long: `Create, update, and delete project labor classifications.

Project labor classifications link projects to labor classifications and
capture hourly rates used for prevailing wage calculations.

Commands:
  create  Create a new project labor classification
  update  Update an existing project labor classification
  delete  Delete a project labor classification`,
	Example: `  # Create a project labor classification
  xbe do project-labor-classifications create --project 123 --labor-classification 456 --basic-hourly-rate 45 --fringe-hourly-rate 12

  # Update hourly rates
  xbe do project-labor-classifications update 789 --basic-hourly-rate 50

  # Delete a project labor classification
  xbe do project-labor-classifications delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectLaborClassificationsCmd)
}
