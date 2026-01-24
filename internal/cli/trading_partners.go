package cli

import "github.com/spf13/cobra"

var tradingPartnersCmd = &cobra.Command{
	Use:   "trading-partners",
	Short: "Browse trading partner links",
	Long: `Browse trading partner links.

Trading partners connect organizations (brokers or customers) with partners
(customers, brokers, truckers, or material sites). Use these commands to list
and inspect trading-partner records.

Commands:
  list    List trading partners with filtering and pagination
  show    Show a trading partner by ID`,
	Example: `  # List trading partner links
  xbe view trading-partners list

  # Filter by organization
  xbe view trading-partners list --organization "Broker|123"

  # Filter by partner
  xbe view trading-partners list --partner "Customer|456"

  # Show a trading partner
  xbe view trading-partners show 789`,
}

func init() {
	viewCmd.AddCommand(tradingPartnersCmd)
}
