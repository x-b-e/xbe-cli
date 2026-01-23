package cli

import "github.com/spf13/cobra"

var customerCommitmentsCmd = &cobra.Command{
	Use:     "customer-commitments",
	Aliases: []string{"customer-commitment"},
	Short:   "Browse customer commitments",
	Long: `Browse customer commitments.

Commands:
  list    List customer commitments with filtering
  show    Show customer commitment details`,
	Example: `  # List customer commitments
  xbe view customer-commitments list

  # Show a customer commitment
  xbe view customer-commitments show 123`,
}

func init() {
	viewCmd.AddCommand(customerCommitmentsCmd)
}
