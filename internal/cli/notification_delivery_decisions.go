package cli

import "github.com/spf13/cobra"

var notificationDeliveryDecisionsCmd = &cobra.Command{
	Use:     "notification-delivery-decisions",
	Aliases: []string{"notification-delivery-decision"},
	Short:   "Browse notification delivery decisions",
	Long: `Browse notification delivery decisions.

Notification delivery decisions capture the evaluated delivery value and timing
for a notification, including channel-specific value scores.

Commands:
  list    List notification delivery decisions
  show    Show notification delivery decision details`,
	Example: `  # List decisions
  xbe view notification-delivery-decisions list

  # Show a decision
  xbe view notification-delivery-decisions show 123

  # Output as JSON
  xbe view notification-delivery-decisions list --json`,
}

func init() {
	viewCmd.AddCommand(notificationDeliveryDecisionsCmd)
}
