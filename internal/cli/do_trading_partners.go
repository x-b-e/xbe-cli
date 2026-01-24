package cli

import "github.com/spf13/cobra"

var doTradingPartnersCmd = &cobra.Command{
	Use:   "trading-partners",
	Short: "Manage trading partners",
	Long: `Manage trading partner links between organizations and partners.

Commands:
  create    Create a trading partner link
  delete    Delete a trading partner link`,
	Example: `  # Create a trading partner link
  xbe do trading-partners create --organization "Broker|123" --partner "Customer|456"

  # Delete a trading partner link (requires --confirm)
  xbe do trading-partners delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doTradingPartnersCmd)
}
