package cli

import "github.com/spf13/cobra"

var businessUnitsCmd = &cobra.Command{
	Use:   "business-units",
	Short: "View business units",
	Long: `View business units on the XBE platform.

Business units are organizational divisions within a company, used for
grouping and reporting on job production plans.

Commands:
  list    List business units`,
	Example: `  # List business units
  xbe view business-units list

  # Search by name
  xbe view business-units list --name "Paving"`,
}

func init() {
	viewCmd.AddCommand(businessUnitsCmd)
}
