package cli

import "github.com/spf13/cobra"

var doProjectSubscriptionsCmd = &cobra.Command{
	Use:   "project-subscriptions",
	Short: "Manage project subscriptions",
	Long:  "Commands for creating, updating, and deleting project subscriptions.",
}

func init() {
	doCmd.AddCommand(doProjectSubscriptionsCmd)
}
