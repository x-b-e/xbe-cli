package cli

import "github.com/spf13/cobra"

var retainerPaymentsCmd = &cobra.Command{
	Use:     "retainer-payments",
	Aliases: []string{"retainer-payment"},
	Short:   "Browse retainer payments",
	Long: `Browse retainer payments.

Retainer payments represent scheduled or recorded payments tied to retainer periods.
They track the payment amount, status, and kind (pre or closing).

Commands:
  list    List retainer payments with filtering and pagination
  show    Show full details of a retainer payment`,
	Example: `  # List retainer payments
  xbe view retainer-payments list

  # Filter by retainer period
  xbe view retainer-payments list --retainer-period 123

  # Filter by status
  xbe view retainer-payments list --status approved

  # Filter by retainer type
  xbe view retainer-payments list --retainer-type BrokerRetainer

  # Filter by buyer
  xbe view retainer-payments list --buyer 456

  # Filter by pay-on date
  xbe view retainer-payments list --pay-on-min 2025-01-15

  # Show retainer payment details
  xbe view retainer-payments show 789`,
}

func init() {
	viewCmd.AddCommand(retainerPaymentsCmd)
}
