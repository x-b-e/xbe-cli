package cli

import "github.com/spf13/cobra"

func newLehmanRobertsApexViewpointTicketExportsShowCmd() *cobra.Command {
	return newGenericShowCmd("lehman-roberts-apex-viewpoint-ticket-exports")
}

func init() {
	lehmanRobertsApexViewpointTicketExportsCmd.AddCommand(newLehmanRobertsApexViewpointTicketExportsShowCmd())
}
