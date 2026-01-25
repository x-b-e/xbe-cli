package cli

import "github.com/spf13/cobra"

var doDeveloperTruckerCertificationMultipliersCmd = &cobra.Command{
	Use:   "developer-trucker-certification-multipliers",
	Short: "Manage developer trucker certification multipliers",
	Long: `Create, update, and delete developer trucker certification multipliers.

Commands:
  create    Create a developer trucker certification multiplier
  update    Update a developer trucker certification multiplier
  delete    Delete a developer trucker certification multiplier`,
	Example: `  # Create a multiplier
  xbe do developer-trucker-certification-multipliers create \
    --developer-trucker-certification 123 \
    --trailer 456 \
    --multiplier 0.85

  # Update a multiplier
  xbe do developer-trucker-certification-multipliers update 789 --multiplier 0.9

  # Delete a multiplier
  xbe do developer-trucker-certification-multipliers delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doDeveloperTruckerCertificationMultipliersCmd)
}
