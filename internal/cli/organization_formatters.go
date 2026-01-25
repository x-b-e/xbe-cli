package cli

import "github.com/spf13/cobra"

var organizationFormattersCmd = &cobra.Command{
	Use:   "organization-formatters",
	Short: "Browse organization formatters",
	Long: `Browse organization formatters.

Organization formatters are JavaScript-based formatter definitions used to
export time sheets, invoices, project actuals, and other organization data.

Commands:
  list    List organization formatters with filtering
  show    Show organization formatter details`,
	Example: `  # List formatters for a broker organization
  xbe view organization-formatters list --organization "Broker|123"

  # Filter by formatter type
  xbe view organization-formatters list --formatter-type TimeSheetsExportFormatter

  # Show formatter details
  xbe view organization-formatters show 456`,
}

func init() {
	viewCmd.AddCommand(organizationFormattersCmd)
}
