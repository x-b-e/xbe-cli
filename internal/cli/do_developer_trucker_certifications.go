package cli

import "github.com/spf13/cobra"

var doDeveloperTruckerCertificationsCmd = &cobra.Command{
	Use:   "developer-trucker-certifications",
	Short: "Manage developer trucker certifications",
	Long: `Manage developer trucker certifications on the XBE platform.

Developer trucker certifications link developers, truckers, and certification
classifications, optionally scoped by start/end dates.

Commands:
  create    Create a developer trucker certification
  update    Update a developer trucker certification
  delete    Delete a developer trucker certification`,
	Example: `  # Create a developer trucker certification
  xbe do developer-trucker-certifications create --developer 123 --trucker 456 --classification 789 --start-on 2024-01-01 --end-on 2024-12-31 --default-multiplier 1.2

  # Update a developer trucker certification
  xbe do developer-trucker-certifications update 321 --classification 654 --default-multiplier 1.35

  # Delete a developer trucker certification
  xbe do developer-trucker-certifications delete 321 --confirm`,
}

func init() {
	doCmd.AddCommand(doDeveloperTruckerCertificationsCmd)
}
