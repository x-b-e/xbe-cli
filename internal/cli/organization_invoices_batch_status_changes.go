package cli

import "github.com/spf13/cobra"

var organizationInvoicesBatchStatusChangesCmd = &cobra.Command{
	Use:     "organization-invoices-batch-status-changes",
	Aliases: []string{"organization-invoices-batch-status-change"},
	Short:   "View organization invoices batch status changes",
	Long: `View organization invoices batch status changes.

Organization invoices batch status changes record state transitions for
organization invoices batches.

Commands:
  list    List organization invoices batch status changes
  show    Show organization invoices batch status change details`,
	Example: `  # List organization invoices batch status changes
  xbe view organization-invoices-batch-status-changes list

  # Show an organization invoices batch status change
  xbe view organization-invoices-batch-status-changes show 123

  # Output JSON
  xbe view organization-invoices-batch-status-changes list --json`,
}

func init() {
	viewCmd.AddCommand(organizationInvoicesBatchStatusChangesCmd)
}
