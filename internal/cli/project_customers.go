package cli

import "github.com/spf13/cobra"

var projectCustomersCmd = &cobra.Command{
	Use:   "project-customers",
	Short: "Browse project customers",
	Long: `Browse project customers on the XBE platform.

Project customers link a project to a customer organization.

Commands:
  list    List project customers with filtering
  show    Show project customer details`,
	Example: `  # List project customers
  xbe view project-customers list

  # Filter by project
  xbe view project-customers list --project 123

  # Filter by customer
  xbe view project-customers list --customer 456

  # Show a specific project customer
  xbe view project-customers show 789`,
}

func init() {
	viewCmd.AddCommand(projectCustomersCmd)
}
