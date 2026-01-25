package cli

import "github.com/spf13/cobra"

func newParkingSitesShowCmd() *cobra.Command {
	return newGenericShowCmd("parking-sites")
}

func init() {
	parkingSitesCmd.AddCommand(newParkingSitesShowCmd())
}
