package cli

import "github.com/spf13/cobra"

var doOrganizationInvoicesBatchInvoiceFailuresCmd = &cobra.Command{
	Use:     "organization-invoices-batch-invoice-failures",
	Aliases: []string{"organization-invoices-batch-invoice-failure"},
	Short:   "Fail organization invoices batch invoices",
	Long: `Fail organization invoices batch invoices.

Organization invoices batch invoice failures mark successful batch invoices as failed.

Commands:
  create    Fail an organization invoices batch invoice`,
}

func init() {
	doCmd.AddCommand(doOrganizationInvoicesBatchInvoiceFailuresCmd)
}
