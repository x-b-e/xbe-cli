package cli

import "github.com/spf13/cobra"

var hosEventsCmd = &cobra.Command{
	Use:     "hos-events",
	Aliases: []string{"hos-event"},
	Short:   "View HOS events",
	Long: `Commands for viewing hours-of-service (HOS) events.

HOS events record driver duty status changes, logins, ELD power events, and
related timing metadata. Use list to browse events and show for full details.`,
	Example: `  # List HOS events
  xbe view hos-events list

  # Filter by driver (user)
  xbe view hos-events list --driver 123

  # View a specific HOS event
  xbe view hos-events show 456`,
}

func init() {
	viewCmd.AddCommand(hosEventsCmd)
}
