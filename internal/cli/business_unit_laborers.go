package cli

import "github.com/spf13/cobra"

var businessUnitLaborersCmd = &cobra.Command{
	Use:   "business-unit-laborers",
	Short: "Browse business unit laborer links",
	Long: `Browse business unit laborer links.

Business unit laborers associate laborers with specific business units.

Commands:
  list    List business unit laborers with filtering and pagination
  show    Show business unit laborer details`,
	Example: `  # List business unit laborer links
  xbe view business-unit-laborers list

  # Filter by business unit
  xbe view business-unit-laborers list --business-unit 123

  # Filter by laborer
  xbe view business-unit-laborers list --laborer 456

  # Show a business unit laborer link
  xbe view business-unit-laborers show 789`,
}

func init() {
	viewCmd.AddCommand(businessUnitLaborersCmd)
}
