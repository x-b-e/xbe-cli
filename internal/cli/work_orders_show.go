package cli

import "github.com/spf13/cobra"

func newWorkOrdersShowCmd() *cobra.Command {
	return newGenericShowCmd("work-orders")
}

func init() {
	workOrdersCmd.AddCommand(newWorkOrdersShowCmd())
}
