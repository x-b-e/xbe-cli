package cli

import "github.com/spf13/cobra"

var profitImprovementSubscriptionsCmd = &cobra.Command{
	Use:   "profit-improvement-subscriptions",
	Short: "Browse profit improvement subscriptions",
	Long: `Browse profit improvement subscriptions on the XBE platform.

Profit improvement subscriptions control notification delivery for updates to
specific profit improvements.

Commands:
  list    List profit improvement subscriptions with filtering
  show    Show profit improvement subscription details`,
	Example: `  # List subscriptions
  xbe view profit-improvement-subscriptions list

  # Filter by profit improvement
  xbe view profit-improvement-subscriptions list --profit-improvement 123

  # Filter by user
  xbe view profit-improvement-subscriptions list --user 456

  # Show a specific subscription
  xbe view profit-improvement-subscriptions show 789`,
}

func init() {
	viewCmd.AddCommand(profitImprovementSubscriptionsCmd)
}
