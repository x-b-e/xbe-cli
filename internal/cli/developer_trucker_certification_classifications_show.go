package cli

import "github.com/spf13/cobra"

func newDeveloperTruckerCertificationClassificationsShowCmd() *cobra.Command {
	return newGenericShowCmd("developer-trucker-certification-classifications")
}

func init() {
	developerTruckerCertificationClassificationsCmd.AddCommand(newDeveloperTruckerCertificationClassificationsShowCmd())
}
