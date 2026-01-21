package cli

import "github.com/spf13/cobra"

var certificationTypesCmd = &cobra.Command{
	Use:   "certification-types",
	Short: "View certification types",
	Long: `View certification types on the XBE platform.

Certification types define the types of certifications that can be tracked
for drivers, truckers, or equipment (e.g., CDL, HAZMAT, DOT medical).

Commands:
  list    List certification types`,
	Example: `  # List certification types
  xbe view certification-types list

  # Filter by what they apply to
  xbe view certification-types list --can-apply-to driver`,
}

func init() {
	viewCmd.AddCommand(certificationTypesCmd)
}
