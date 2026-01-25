package cli

import "github.com/spf13/cobra"

var doProjectCustomersCmd = &cobra.Command{
	Use:   "project-customers",
	Short: "Manage project customers",
	Long: `Manage project customers on the XBE platform.

Project customers link a project to a customer organization.

Commands:
  create    Create a new project customer
  delete    Delete a project customer`,
	Example: `  # Create a project customer
  xbe do project-customers create --project 123 --customer 456

  # Delete a project customer (requires --confirm)
  xbe do project-customers delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectCustomersCmd)
}
