package cli

import "github.com/spf13/cobra"

var doMeetingsCmd = &cobra.Command{
	Use:     "meetings",
	Aliases: []string{"meeting"},
	Short:   "Manage meetings",
	Long: `Create, update, and delete meetings.

Meetings track scheduled discussions tied to an organization, with optional
organizer, attendees, and related action items.

Actions:
  create  Create a new meeting
  update  Update an existing meeting
  delete  Delete a meeting`,
}

func init() {
	doCmd.AddCommand(doMeetingsCmd)
}
