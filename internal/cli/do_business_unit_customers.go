package cli

import "github.com/spf13/cobra"

var doBusinessUnitCustomersCmd = &cobra.Command{
	Use:   "business-unit-customers",
	Short: "Manage business unit customer links",
	Long: `Create and delete business unit customer links.

Commands:
  create    Create a business unit customer link
  delete    Delete a business unit customer link`,
	Example: `  # Create a business unit customer link
  xbe do business-unit-customers create --business-unit 123 --customer 456

  # Delete a business unit customer link
  xbe do business-unit-customers delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doBusinessUnitCustomersCmd)
}
