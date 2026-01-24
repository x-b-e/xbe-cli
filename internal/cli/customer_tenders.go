package cli

import "github.com/spf13/cobra"

var customerTendersCmd = &cobra.Command{
	Use:     "customer-tenders",
	Aliases: []string{"customer-tender"},
	Short:   "Browse customer tenders",
	Long: `Browse customer tenders on the XBE platform.

Customer tenders are offers from customers to brokers for job work.

Commands:
  list    List customer tenders with filtering
  show    Show customer tender details`,
	Example: `  # List customer tenders
  xbe view customer-tenders list

  # Show a customer tender
  xbe view customer-tenders show 123`,
}

func init() {
	viewCmd.AddCommand(customerTendersCmd)
}
