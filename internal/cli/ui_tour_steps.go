package cli

import "github.com/spf13/cobra"

var uiTourStepsCmd = &cobra.Command{
	Use:   "ui-tour-steps",
	Short: "Browse and view UI tour steps",
	Long: `Browse and view UI tour steps on the XBE platform.

UI tour steps define the ordered prompts and content shown to users
throughout guided product walkthroughs.

Commands:
  list    List UI tour steps with filtering
  show    View UI tour step details`,
	Example: `  # List UI tour steps
  xbe view ui-tour-steps list

  # View a specific step
  xbe view ui-tour-steps show 123`,
}

func init() {
	viewCmd.AddCommand(uiTourStepsCmd)
}
