package cli

import "github.com/spf13/cobra"

func newSuperiorBowenApexViewpointTicketExportsShowCmd() *cobra.Command {
	return newGenericShowCmd("superior-bowen-apex-viewpoint-ticket-exports")
}

func init() {
	superiorBowenApexViewpointTicketExportsCmd.AddCommand(newSuperiorBowenApexViewpointTicketExportsShowCmd())
}
