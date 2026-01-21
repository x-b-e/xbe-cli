package cli

import "github.com/spf13/cobra"

var costIndexEntriesCmd = &cobra.Command{
	Use:   "cost-index-entries",
	Short: "View cost index entries",
	Long: `View cost index entries.

Cost index entries are time-series values for cost indexes, used in rate adjustments.

Commands:
  list  List cost index entries`,
}

func init() {
	viewCmd.AddCommand(costIndexEntriesCmd)
}
