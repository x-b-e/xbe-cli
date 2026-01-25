package cli

import "github.com/spf13/cobra"

var developerTruckerCertificationsCmd = &cobra.Command{
	Use:   "developer-trucker-certifications",
	Short: "View developer trucker certifications",
	Long: `View developer trucker certifications.

Developer trucker certifications link developers, truckers, and certification
classifications, optionally scoped by start/end dates.

Commands:
  list  List developer trucker certifications
  show  Show developer trucker certification details`,
}

func init() {
	viewCmd.AddCommand(developerTruckerCertificationsCmd)
}
