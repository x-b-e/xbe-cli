package cli

import "github.com/spf13/cobra"

var doRateAgreementsCmd = &cobra.Command{
	Use:     "rate-agreements",
	Aliases: []string{"rate-agreement"},
	Short:   "Manage rate agreements",
	Long: `Create, update, and delete rate agreements.

Rate agreements define negotiated pricing between sellers (brokers or truckers)
and buyers (customers or brokers).

Commands:
  create    Create a new rate agreement
  update    Update an existing rate agreement
  delete    Delete a rate agreement`,
	Example: `  # Create a rate agreement
  xbe do rate-agreements create --name "Standard" --status active --seller "Broker|123" --buyer "Customer|456"

  # Update a rate agreement
  xbe do rate-agreements update 789 --status inactive

  # Delete a rate agreement
  xbe do rate-agreements delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doRateAgreementsCmd)
}
