package cli

import "github.com/spf13/cobra"

var openDoorIssuesCmd = &cobra.Command{
	Use:   "open-door-issues",
	Short: "View open door issues",
	Long: `View open door issues reported for broker, customer, or trucker organizations.

Open door issues capture concerns submitted by users and tracked to resolution.
Use the do commands to create, update, or delete issues.

Commands:
  list    List open door issues
  show    Show open door issue details`,
	Example: `  # List open door issues
  xbe view open-door-issues list

  # Filter by organization
  xbe view open-door-issues list --organization "Broker|123"

  # Show details for an issue
  xbe view open-door-issues show 456`,
}

func init() {
	viewCmd.AddCommand(openDoorIssuesCmd)
}
