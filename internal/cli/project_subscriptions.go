package cli

import "github.com/spf13/cobra"

var projectSubscriptionsCmd = &cobra.Command{
	Use:   "project-subscriptions",
	Short: "View project subscriptions",
	Long: `View project subscriptions on the XBE platform.

Project subscriptions determine which users receive
notifications for a specific project.

Commands:
  list    List project subscriptions
  show    Show project subscription details`,
	Example: `  # List subscriptions
  xbe view project-subscriptions list

  # Show a subscription
  xbe view project-subscriptions show 123

  # Output as JSON
  xbe view project-subscriptions list --json`,
}

func init() {
	viewCmd.AddCommand(projectSubscriptionsCmd)
}
