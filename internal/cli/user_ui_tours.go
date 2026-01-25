package cli

import "github.com/spf13/cobra"

var userUiToursCmd = &cobra.Command{
	Use:   "user-ui-tours",
	Short: "Browse user UI tours",
	Long: `Browse user UI tours on the XBE platform.

User UI tours capture when a user completes or skips a UI tour.

Commands:
  list    List user UI tours with filtering
  show    View user UI tour details`,
	Example: `  # List user UI tours
  xbe view user-ui-tours list

  # Show a user UI tour
  xbe view user-ui-tours show 123`,
}

func init() {
	viewCmd.AddCommand(userUiToursCmd)
}
