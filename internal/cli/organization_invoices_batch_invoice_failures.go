package cli

import "github.com/spf13/cobra"

var organizationInvoicesBatchInvoiceFailuresCmd = &cobra.Command{
	Use:     "organization-invoices-batch-invoice-failures",
	Aliases: []string{"organization-invoices-batch-invoice-failure"},
	Short:   "View organization invoices batch invoice failures",
	Long: `View organization invoices batch invoice failures.

Organization invoices batch invoice failures mark successful batch invoices as failed.

Commands:
  list    List organization invoices batch invoice failures
  show    Show organization invoices batch invoice failure details`,
	Example: `  # List organization invoices batch invoice failures
  xbe view organization-invoices-batch-invoice-failures list

  # Show an organization invoices batch invoice failure
  xbe view organization-invoices-batch-invoice-failures show 123

  # Output JSON
  xbe view organization-invoices-batch-invoice-failures list --json`,
}

func init() {
	viewCmd.AddCommand(organizationInvoicesBatchInvoiceFailuresCmd)
}
