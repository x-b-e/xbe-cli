package cli

import "github.com/spf13/cobra"

var materialSiteSubscriptionsCmd = &cobra.Command{
	Use:   "material-site-subscriptions",
	Short: "Browse material site subscriptions",
	Long: `Browse material site subscriptions on the XBE platform.

Material site subscriptions control notification delivery for activity at
specific material sites.

Commands:
  list    List material site subscriptions with filtering
  show    Show material site subscription details`,
	Example: `  # List subscriptions
  xbe view material-site-subscriptions list

  # Filter by material site
  xbe view material-site-subscriptions list --material-site 123

  # Filter by user
  xbe view material-site-subscriptions list --user 456

  # Show a specific subscription
  xbe view material-site-subscriptions show 789`,
}

func init() {
	viewCmd.AddCommand(materialSiteSubscriptionsCmd)
}
