package cli

import "github.com/spf13/cobra"

var organizationInvoicesBatchFilesCmd = &cobra.Command{
	Use:   "organization-invoices-batch-files",
	Short: "Browse organization invoices batch files",
	Long: `Browse organization invoices batch files on the XBE platform.

Organization invoices batch files represent formatted invoice batch outputs
produced by organization formatters.

Commands:
  list    List organization invoices batch files with filtering
  show    Show organization invoices batch file details`,
	Example: `  # List organization invoices batch files
  xbe view organization-invoices-batch-files list

  # Show an organization invoices batch file
  xbe view organization-invoices-batch-files show 123`,
}

func init() {
	viewCmd.AddCommand(organizationInvoicesBatchFilesCmd)
}
