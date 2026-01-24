package cli

import "github.com/spf13/cobra"

var customerApplicationsCmd = &cobra.Command{
	Use:     "customer-applications",
	Aliases: []string{"customer-application"},
	Short:   "Browse customer applications",
	Long: `Browse customer applications.

Commands:
  list    List customer applications with filtering
  show    Show customer application details`,
	Example: `  # List customer applications
  xbe view customer-applications list

  # Show a customer application
  xbe view customer-applications show 123`,
}

func init() {
	viewCmd.AddCommand(customerApplicationsCmd)
}
