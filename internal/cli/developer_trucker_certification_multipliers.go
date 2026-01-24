package cli

import "github.com/spf13/cobra"

var developerTruckerCertificationMultipliersCmd = &cobra.Command{
	Use:   "developer-trucker-certification-multipliers",
	Short: "Browse developer trucker certification multipliers",
	Long: `Browse developer trucker certification multipliers.

Developer trucker certification multipliers set trailer-specific multipliers
for a developer trucker certification.

Commands:
  list    List multipliers with filtering and pagination
  show    Show multiplier details`,
	Example: `  # List multipliers
  xbe view developer-trucker-certification-multipliers list

  # Filter by developer trucker certification
  xbe view developer-trucker-certification-multipliers list --developer-trucker-certification 123

  # Filter by trailer
  xbe view developer-trucker-certification-multipliers list --trailer 456

  # Show multiplier details
  xbe view developer-trucker-certification-multipliers show 789`,
}

func init() {
	viewCmd.AddCommand(developerTruckerCertificationMultipliersCmd)
}
