package cli

import "github.com/spf13/cobra"

var doMaterialSiteSubscriptionsCmd = &cobra.Command{
	Use:   "material-site-subscriptions",
	Short: "Manage material site subscriptions",
	Long: `Manage material site subscriptions on the XBE platform.

Material site subscriptions define which users receive notifications for
activity at a specific material site.

Commands:
  create    Create a new material site subscription
  update    Update a material site subscription
  delete    Delete a material site subscription`,
	Example: `  # Create a subscription
  xbe do material-site-subscriptions create --user 123 --material-site 456 --contact-method email_address

  # Update a subscription
  xbe do material-site-subscriptions update 789 --contact-method mobile_number

  # Delete a subscription (requires --confirm)
  xbe do material-site-subscriptions delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doMaterialSiteSubscriptionsCmd)
}
