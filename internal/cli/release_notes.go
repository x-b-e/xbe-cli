package cli

import "github.com/spf13/cobra"

var releaseNotesCmd = &cobra.Command{
	Use:   "release-notes",
	Short: "Browse and view release notes",
	Long: `Browse and view release notes on the XBE platform.

Release notes document product updates, new features, and improvements.
Each release note includes a headline, description, and release date.

Commands:
  list    List release notes with filtering and pagination
  show    View the full details of a specific release note`,
	Example: `  # List recent release notes
  xbe view release-notes list

  # Filter by published status
  xbe view release-notes list --is-published true

  # Search release notes
  xbe view release-notes list --q "trucking"

  # View a specific release note
  xbe view release-notes show 123`,
}

func init() {
	viewCmd.AddCommand(releaseNotesCmd)
}
