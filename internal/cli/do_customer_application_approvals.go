package cli

import "github.com/spf13/cobra"

var doCustomerApplicationApprovalsCmd = &cobra.Command{
	Use:   "customer-application-approvals",
	Short: "Approve customer applications",
	Long: `Approve customer applications.

Customer application approvals create a customer from the application data
and mark the application as approved.

Commands:
  create    Approve a customer application`,
	Example: `  # Approve a customer application
  xbe do customer-application-approvals create --customer-application 123 --credit-limit 1000000

  # JSON output
  xbe do customer-application-approvals create --customer-application 123 --credit-limit 1000000 --json`,
}

func init() {
	doCmd.AddCommand(doCustomerApplicationApprovalsCmd)
}
