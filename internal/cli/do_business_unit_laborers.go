package cli

import "github.com/spf13/cobra"

var doBusinessUnitLaborersCmd = &cobra.Command{
	Use:   "business-unit-laborers",
	Short: "Manage business unit laborer links",
	Long: `Create and delete business unit laborer links.

Commands:
  create    Create a business unit laborer link
  delete    Delete a business unit laborer link`,
	Example: `  # Create a business unit laborer link
  xbe do business-unit-laborers create --business-unit 123 --laborer 456

  # Delete a business unit laborer link
  xbe do business-unit-laborers delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doBusinessUnitLaborersCmd)
}
