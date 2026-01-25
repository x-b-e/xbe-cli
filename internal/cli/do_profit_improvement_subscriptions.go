package cli

import "github.com/spf13/cobra"

var doProfitImprovementSubscriptionsCmd = &cobra.Command{
	Use:   "profit-improvement-subscriptions",
	Short: "Manage profit improvement subscriptions",
	Long: `Manage profit improvement subscriptions on the XBE platform.

Profit improvement subscriptions define which users receive notifications
for updates to a profit improvement.

Commands:
  create    Create a new profit improvement subscription
  update    Update a profit improvement subscription
  delete    Delete a profit improvement subscription`,
	Example: `  # Create a subscription
  xbe do profit-improvement-subscriptions create --user 123 --profit-improvement 456 --contact-method email_address

  # Update a subscription
  xbe do profit-improvement-subscriptions update 789 --contact-method mobile_number

  # Delete a subscription (requires --confirm)
  xbe do profit-improvement-subscriptions delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doProfitImprovementSubscriptionsCmd)
}
