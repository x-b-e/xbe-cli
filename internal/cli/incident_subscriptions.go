package cli

import "github.com/spf13/cobra"

var incidentSubscriptionsCmd = &cobra.Command{
	Use:     "incident-subscriptions",
	Aliases: []string{"incident-subscription"},
	Short:   "View incident subscriptions",
	Long: `View incident subscriptions.

Incident subscriptions notify users about incidents scoped to an organization
or a specific incident. Optional kind filters can further scope notifications.

Commands:
  list    List incident subscriptions
  show    Show incident subscription details`,
	Example: `  # List incident subscriptions
  xbe view incident-subscriptions list

  # Filter by user and broker
  xbe view incident-subscriptions list --user 123 --broker 456

  # Show a subscription
  xbe view incident-subscriptions show 123`,
}

func init() {
	viewCmd.AddCommand(incidentSubscriptionsCmd)
}
