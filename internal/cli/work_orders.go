package cli

import "github.com/spf13/cobra"

var workOrdersCmd = &cobra.Command{
	Use:     "work-orders",
	Aliases: []string{"work-order"},
	Short:   "View work orders",
	Long:    "Commands for viewing work orders.",
}

func init() {
	viewCmd.AddCommand(workOrdersCmd)
}
