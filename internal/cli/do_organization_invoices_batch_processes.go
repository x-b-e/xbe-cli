package cli

import "github.com/spf13/cobra"

var doOrganizationInvoicesBatchProcessesCmd = &cobra.Command{
	Use:     "organization-invoices-batch-processes",
	Aliases: []string{"organization-invoices-batch-process"},
	Short:   "Process organization invoices batches",
	Long: `Process organization invoices batches.

Organization invoices batch processes mark batches as processed.

Commands:
  create    Process an organization invoices batch`,
}

func init() {
	doCmd.AddCommand(doOrganizationInvoicesBatchProcessesCmd)
}
