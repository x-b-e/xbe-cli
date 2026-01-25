package cli

import "github.com/spf13/cobra"

var tenderJobScheduleShiftsMaterialTransactionsChecksumsCmd = &cobra.Command{
	Use:   "tender-job-schedule-shifts-material-transactions-checksums",
	Short: "Browse tender job schedule shift material transaction checksums",
	Long: `Browse tender job schedule shift material transaction checksums.

Checksum records compare raw material transactions to tender job schedule shifts
for a given job number and time window.

Commands:
  list  List checksum records
  show  Show checksum details`,
	Example: `  # List checksum records
  xbe view tender-job-schedule-shifts-material-transactions-checksums list

  # Show checksum details
  xbe view tender-job-schedule-shifts-material-transactions-checksums show 123`,
}

func init() {
	viewCmd.AddCommand(tenderJobScheduleShiftsMaterialTransactionsChecksumsCmd)
}
