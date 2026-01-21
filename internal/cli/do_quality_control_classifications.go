package cli

import "github.com/spf13/cobra"

var doQualityControlClassificationsCmd = &cobra.Command{
	Use:   "quality-control-classifications",
	Short: "Manage quality control classifications",
	Long: `Create, update, and delete quality control classifications.

Quality control classifications define types of quality inspections
and checks that can be performed, scoped to a broker organization.

Commands:
  create    Create a new quality control classification
  update    Update an existing quality control classification
  delete    Delete a quality control classification`,
	Example: `  # Create a quality control classification
  xbe do quality-control-classifications create --name "Temperature Check" --broker 123

  # Update a quality control classification
  xbe do quality-control-classifications update 456 --description "Check material temperature"

  # Delete a quality control classification
  xbe do quality-control-classifications delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doQualityControlClassificationsCmd)
}
