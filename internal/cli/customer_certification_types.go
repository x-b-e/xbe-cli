package cli

import "github.com/spf13/cobra"

var customerCertificationTypesCmd = &cobra.Command{
	Use:   "customer-certification-types",
	Short: "View customer certification types",
	Long: `View customer certification types on the XBE platform.

Customer certification types link customers to certification types they can
track or require.

Commands:
  list    List customer certification types
  show    Show customer certification type details`,
	Example: `  # List customer certification types
  xbe view customer-certification-types list

  # Show a customer certification type
  xbe view customer-certification-types show 123`,
}

func init() {
	viewCmd.AddCommand(customerCertificationTypesCmd)
}
