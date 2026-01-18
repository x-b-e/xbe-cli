package cli

import "github.com/spf13/cobra"

var pressReleasesCmd = &cobra.Command{
	Use:   "press-releases",
	Short: "Browse and view press releases",
	Long: `Browse and view press releases on the XBE platform.

Press releases are official announcements about company news, product launches,
partnerships, and other significant events.

Commands:
  list    List press releases
  show    View the full details of a specific press release`,
	Example: `  # List all press releases
  xbe view press-releases list

  # Filter by published status
  xbe view press-releases list --published true

  # View a specific press release
  xbe view press-releases show 123`,
}

func init() {
	viewCmd.AddCommand(pressReleasesCmd)
}
