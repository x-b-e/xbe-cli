package cli

import "github.com/spf13/cobra"

var doOrganizationInvoicesBatchFileExportsCmd = &cobra.Command{
	Use:     "organization-invoices-batch-file-exports",
	Aliases: []string{"organization-invoices-batch-file-export"},
	Short:   "Export organization invoices batch files",
	Long: `Export organization invoices batch files.

Organization invoices batch file exports send formatted invoice batch files to
partner integrations.

Commands:
  create    Export an organization invoices batch file`,
	Example: `  # Export a batch file
  xbe do organization-invoices-batch-file-exports create --organization-invoices-batch-file 123

  # Run export as a dry run
  xbe do organization-invoices-batch-file-exports create --organization-invoices-batch-file 123 --dry-run`,
}

func init() {
	doCmd.AddCommand(doOrganizationInvoicesBatchFileExportsCmd)
}
