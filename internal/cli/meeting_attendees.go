package cli

import "github.com/spf13/cobra"

var meetingAttendeesCmd = &cobra.Command{
	Use:   "meeting-attendees",
	Short: "Browse meeting attendees",
	Long: `Browse meeting attendees.

Meeting attendees link users to meetings and track attendance status.

Commands:
  list    List meeting attendees with filtering and pagination
  show    Show meeting attendee details`,
	Example: `  # List meeting attendees
  xbe view meeting-attendees list

  # Filter by meeting
  xbe view meeting-attendees list --meeting 123

  # Show a meeting attendee
  xbe view meeting-attendees show 456`,
}

func init() {
	viewCmd.AddCommand(meetingAttendeesCmd)
}
