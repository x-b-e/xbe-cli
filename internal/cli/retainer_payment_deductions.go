package cli

import "github.com/spf13/cobra"

var retainerPaymentDeductionsCmd = &cobra.Command{
	Use:     "retainer-payment-deductions",
	Aliases: []string{"retainer-payment-deduction"},
	Short:   "View retainer payment deductions",
	Long: `Browse retainer payment deductions.

Retainer payment deductions record the applied amount from a retainer payment to a
retainer deduction. Use list to browse and show for full details.`,
	Example: `  # List retainer payment deductions
  xbe view retainer-payment-deductions list

  # Filter by created date
  xbe view retainer-payment-deductions list --created-at-min 2025-01-01T00:00:00Z

  # View a retainer payment deduction
  xbe view retainer-payment-deductions show 123`,
}

func init() {
	viewCmd.AddCommand(retainerPaymentDeductionsCmd)
}
