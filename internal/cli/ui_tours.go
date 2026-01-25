package cli

import "github.com/spf13/cobra"

var uiToursCmd = &cobra.Command{
	Use:   "ui-tours",
	Short: "View UI tours",
	Long: `View UI tours on the XBE platform.

UI tours define guided walkthroughs for users across the web app.

Commands:
  list    List UI tours
  show    Show UI tour details`,
	Example: `  # List UI tours
  xbe view ui-tours list

  # Filter by abbreviation
  xbe view ui-tours list --abbreviation "onboarding"

  # Show a UI tour
  xbe view ui-tours show 123`,
}

func init() {
	viewCmd.AddCommand(uiToursCmd)
}
