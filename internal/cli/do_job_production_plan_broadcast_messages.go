package cli

import "github.com/spf13/cobra"

var doJobProductionPlanBroadcastMessagesCmd = &cobra.Command{
	Use:   "job-production-plan-broadcast-messages",
	Short: "Manage job production plan broadcast messages",
	Long: `Create and update job production plan broadcast messages.

Broadcast messages notify job production plan participants and can be hidden
without deleting the record.

Commands:
  create    Create a broadcast message
  update    Update a broadcast message`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanBroadcastMessagesCmd)
}
