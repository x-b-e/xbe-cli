package cli

import "github.com/spf13/cobra"

var doMeetingAttendeesCmd = &cobra.Command{
	Use:   "meeting-attendees",
	Short: "Manage meeting attendees",
	Long: `Create, update, and delete meeting attendees.

Commands:
  create    Create a meeting attendee
  update    Update a meeting attendee
  delete    Delete a meeting attendee`,
	Example: `  # Create a meeting attendee
  xbe do meeting-attendees create --meeting 123 --user 456 --location-kind on_site

  # Update attendance
  xbe do meeting-attendees update 789 --is-present true

  # Delete a meeting attendee
  xbe do meeting-attendees delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doMeetingAttendeesCmd)
}
