package cli

import "github.com/spf13/cobra"

var userLocationRequestsCmd = &cobra.Command{
	Use:     "user-location-requests",
	Aliases: []string{"user-location-request"},
	Short:   "Browse user location requests",
	Long: `Browse user location requests.

User location requests ask a user to share their current location.

Commands:
  list    List user location requests with filtering and pagination
  show    Show user location request details`,
	Example: `  # List user location requests
  xbe view user-location-requests list

  # Show a user location request
  xbe view user-location-requests show 123`,
}

func init() {
	viewCmd.AddCommand(userLocationRequestsCmd)
}
