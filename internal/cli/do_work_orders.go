package cli

import "github.com/spf13/cobra"

var doWorkOrdersCmd = &cobra.Command{
	Use:     "work-orders",
	Aliases: []string{"work-order"},
	Short:   "Manage work orders",
	Long:    "Commands for creating, updating, and deleting work orders.",
}

func init() {
	doCmd.AddCommand(doWorkOrdersCmd)
}
