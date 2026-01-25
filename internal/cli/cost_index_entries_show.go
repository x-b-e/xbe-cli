package cli

import "github.com/spf13/cobra"

func newCostIndexEntriesShowCmd() *cobra.Command {
	return newGenericShowCmd("cost-index-entries")
}

func init() {
	costIndexEntriesCmd.AddCommand(newCostIndexEntriesShowCmd())
}
