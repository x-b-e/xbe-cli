package cli

import "github.com/spf13/cobra"

var doTenderJobScheduleShiftsMaterialTransactionsChecksumsCmd = &cobra.Command{
	Use:   "tender-job-schedule-shifts-material-transactions-checksums",
	Short: "Generate tender job schedule shift material transaction checksums",
	Long:  "Commands for generating checksum diagnostics between shifts and material transactions.",
}

func init() {
	doCmd.AddCommand(doTenderJobScheduleShiftsMaterialTransactionsChecksumsCmd)
}
