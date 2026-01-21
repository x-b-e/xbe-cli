package cli

import "github.com/spf13/cobra"

var doDeveloperTruckerCertificationClassificationsCmd = &cobra.Command{
	Use:   "developer-trucker-certification-classifications",
	Short: "Manage developer trucker certification classifications",
	Long: `Create, update, and delete developer trucker certification classifications.

These classifications define types of certifications that truckers can have for a developer.

Commands:
  create  Create a new developer trucker certification classification
  update  Update an existing developer trucker certification classification
  delete  Delete a developer trucker certification classification`,
	Example: `  # Create a developer trucker certification classification
  xbe do developer-trucker-certification-classifications create --name "Safety Training" --developer 123

  # Update a developer trucker certification classification
  xbe do developer-trucker-certification-classifications update 456 --name "Updated Name"

  # Delete a developer trucker certification classification
  xbe do developer-trucker-certification-classifications delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doDeveloperTruckerCertificationClassificationsCmd)
}
