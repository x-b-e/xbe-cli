package cli

import "github.com/spf13/cobra"

var doOzingaTkBatchFileExportsCmd = &cobra.Command{
	Use:     "ozinga-tk-batch-file-exports",
	Aliases: []string{"ozinga-tk-batch-file-export"},
	Short:   "Generate Ozinga TK batch file exports",
	Long: `Generate Ozinga TK batch file exports.

Exports submit a processed organization invoices batch file to the Ozinga TK
integration and return export results or errors.

Commands:
  create    Create an export`,
	Example: `  # Create an export
  xbe do ozinga-tk-batch-file-exports create --organization-invoices-batch-file 123

  # Dry-run export
  xbe do ozinga-tk-batch-file-exports create --organization-invoices-batch-file 123 --dry-run`,
}

func init() {
	doCmd.AddCommand(doOzingaTkBatchFileExportsCmd)
}
