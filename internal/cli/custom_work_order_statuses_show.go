package cli

import "github.com/spf13/cobra"

func newCustomWorkOrderStatusesShowCmd() *cobra.Command {
	return newGenericShowCmd("custom-work-order-statuses")
}

func init() {
	customWorkOrderStatusesCmd.AddCommand(newCustomWorkOrderStatusesShowCmd())
}
