package cli

import "github.com/spf13/cobra"

var doLaborClassificationsCmd = &cobra.Command{
	Use:   "labor-classifications",
	Short: "Manage labor classifications",
	Long: `Manage labor classifications on the XBE platform.

Labor classifications define types of workers (e.g., raker, screedman, foreman)
with their capabilities and permissions.

Commands:
  create    Create a new labor classification
  update    Update an existing labor classification
  delete    Delete a labor classification`,
	Example: `  # Create a labor classification
  xbe do labor-classifications create --name "Raker" --abbreviation "raker"

  # Update a labor classification
  xbe do labor-classifications update 123 --name "Senior Raker"

  # Delete a labor classification (requires --confirm)
  xbe do labor-classifications delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doLaborClassificationsCmd)
}
