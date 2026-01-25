package cli

import "github.com/spf13/cobra"

var integrationInvoicesBatchExportsCmd = &cobra.Command{
	Use:   "integration-invoices-batch-exports",
	Short: "Browse integration invoices batch exports",
	Long: `Browse integration invoices batch exports on the XBE platform.

Integration invoices batch exports track the exports produced for integration
targets alongside their invoice batch context.

Commands:
  list    List integration invoices batch exports with filtering
  show    Show integration invoices batch export details`,
	Example: `  # List integration invoices batch exports
  xbe view integration-invoices-batch-exports list

  # Filter by organization invoices batch
  xbe view integration-invoices-batch-exports list --organization-invoices-batch 123

  # Show an integration invoices batch export
  xbe view integration-invoices-batch-exports show 456`,
}

func init() {
	viewCmd.AddCommand(integrationInvoicesBatchExportsCmd)
}
