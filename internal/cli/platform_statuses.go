package cli

import "github.com/spf13/cobra"

var platformStatusesCmd = &cobra.Command{
	Use:   "platform-statuses",
	Short: "Browse platform status updates",
	Long: `Browse platform status updates on the XBE platform.

Platform statuses communicate incidents, maintenance windows, and other
service updates visible to platform users.

Commands:
  list    List platform statuses
  show    View platform status details`,
	Example: `  # List platform statuses
  xbe view platform-statuses list

  # View a specific platform status
  xbe view platform-statuses show 123`,
}

func init() {
	viewCmd.AddCommand(platformStatusesCmd)
}
