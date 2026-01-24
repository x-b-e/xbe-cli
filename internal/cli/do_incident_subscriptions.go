package cli

import "github.com/spf13/cobra"

var doIncidentSubscriptionsCmd = &cobra.Command{
	Use:     "incident-subscriptions",
	Aliases: []string{"incident-subscription"},
	Short:   "Manage incident subscriptions",
	Long: `Create, update, and delete incident subscriptions.

Incident subscriptions notify users about incidents scoped to an organization
or a specific incident. Non-admin users must scope subscriptions to an
organization or an incident.`,
	Example: `  # Create a subscription for a broker
  xbe do incident-subscriptions create --user 123 --broker 456 --kind safety

  # Update contact method
  xbe do incident-subscriptions update 789 --contact-method mobile_number

  # Delete a subscription
  xbe do incident-subscriptions delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doIncidentSubscriptionsCmd)
}
