package cli

import "github.com/spf13/cobra"

var doOpenDoorIssuesCmd = &cobra.Command{
	Use:   "open-door-issues",
	Short: "Manage open door issues",
	Long: `Create, update, and delete open door issues.

Open door issues capture concerns reported by users for broker, customer,
or trucker organizations.`,
	Example: `  # Create an open door issue
  xbe do open-door-issues create \
    --description "Driver reported a safety concern" \
    --status editing \
    --organization "Broker|123" \
    --reported-by 456

  # Update an open door issue
  xbe do open-door-issues update 789 --status resolved

  # Delete an open door issue
  xbe do open-door-issues delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doOpenDoorIssuesCmd)
}
