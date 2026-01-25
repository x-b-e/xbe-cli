package cli

import "github.com/spf13/cobra"

func newTimeSheetLineItemClassificationsShowCmd() *cobra.Command {
	return newGenericShowCmd("time-sheet-line-item-classifications")
}

func init() {
	timeSheetLineItemClassificationsCmd.AddCommand(newTimeSheetLineItemClassificationsShowCmd())
}
