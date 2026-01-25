package cli

import "github.com/spf13/cobra"

var doOrganizationFormattersCmd = &cobra.Command{
	Use:   "organization-formatters",
	Short: "Manage organization formatters",
	Long: `Manage organization formatters.

Organization formatters are JavaScript formatter definitions used to export
time sheets, invoices, and other organization-specific data.

Commands:
  create    Create a new organization formatter
  update    Update an existing organization formatter`,
	Example: `  # Create a formatter
  xbe do organization-formatters create --formatter-type TimeSheetsExportFormatter \\
    --organization Broker|123 \\
    --formatter-function 'function format(lineItemsJson, timestamp) { return lineItemsJson; }'

  # Update formatter description
  xbe do organization-formatters update 456 --description \"Updated formatter\"`,
}

func init() {
	doCmd.AddCommand(doOrganizationFormattersCmd)
}
