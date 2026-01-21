package cli

import "github.com/spf13/cobra"

var developerTruckerCertificationClassificationsCmd = &cobra.Command{
	Use:   "developer-trucker-certification-classifications",
	Short: "View developer trucker certification classifications",
	Long: `View developer trucker certification classifications.

These classifications define types of certifications that truckers can have for a developer.

Commands:
  list  List developer trucker certification classifications`,
}

func init() {
	viewCmd.AddCommand(developerTruckerCertificationClassificationsCmd)
}
