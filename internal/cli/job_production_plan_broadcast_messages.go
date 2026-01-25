package cli

import "github.com/spf13/cobra"

var jobProductionPlanBroadcastMessagesCmd = &cobra.Command{
	Use:   "job-production-plan-broadcast-messages",
	Short: "Browse job production plan broadcast messages",
	Long: `Browse job production plan broadcast messages on the XBE platform.

Broadcast messages notify participants on a job production plan about
schedule updates, logistics, or important changes.

Commands:
  list    List broadcast messages with filtering
  show    Show details of a specific broadcast message`,
	Example: `  # List broadcast messages for a job production plan
  xbe view job-production-plan-broadcast-messages list --job-production-plan 123

  # Include hidden messages
  xbe view job-production-plan-broadcast-messages list --job-production-plan 123 --is-hidden true

  # Show message details
  xbe view job-production-plan-broadcast-messages show 456`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanBroadcastMessagesCmd)
}
