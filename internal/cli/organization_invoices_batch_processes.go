package cli

import "github.com/spf13/cobra"

var organizationInvoicesBatchProcessesCmd = &cobra.Command{
	Use:     "organization-invoices-batch-processes",
	Aliases: []string{"organization-invoices-batch-process"},
	Short:   "View organization invoices batch processes",
	Long: `View organization invoices batch processes.

Organization invoices batch processes transition batches from not processed to processed.

Commands:
  list    List organization invoices batch processes
  show    Show organization invoices batch process details`,
	Example: `  # List organization invoices batch processes
  xbe view organization-invoices-batch-processes list

  # Show an organization invoices batch process
  xbe view organization-invoices-batch-processes show 123

  # Output JSON
  xbe view organization-invoices-batch-processes list --json`,
}

func init() {
	viewCmd.AddCommand(organizationInvoicesBatchProcessesCmd)
}
