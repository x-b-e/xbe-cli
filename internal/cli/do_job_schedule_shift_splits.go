package cli

import "github.com/spf13/cobra"

var doJobScheduleShiftSplitsCmd = &cobra.Command{
	Use:     "job-schedule-shift-splits",
	Aliases: []string{"job-schedule-shift-split"},
	Short:   "Split job schedule shifts",
	Long:    "Commands for splitting job schedule shifts.",
}

func init() {
	doCmd.AddCommand(doJobScheduleShiftSplitsCmd)
}
