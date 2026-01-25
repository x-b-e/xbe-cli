package cli

import "github.com/spf13/cobra"

var doUiToursCmd = &cobra.Command{
	Use:   "ui-tours",
	Short: "Manage UI tours",
	Long: `Manage UI tours on the XBE platform.

UI tours define guided walkthroughs for users across the web app.

Commands:
  create    Create a new UI tour
  update    Update an existing UI tour
  delete    Delete a UI tour`,
	Example: `  # Create a UI tour
  xbe do ui-tours create --name "Driver Onboarding" --abbreviation "driver-onboarding"

  # Update a UI tour
  xbe do ui-tours update 123 --description "Updated walkthrough"

  # Delete a UI tour (requires --confirm)
  xbe do ui-tours delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doUiToursCmd)
}
