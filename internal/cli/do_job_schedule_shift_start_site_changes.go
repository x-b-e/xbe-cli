package cli

import "github.com/spf13/cobra"

var doJobScheduleShiftStartSiteChangesCmd = &cobra.Command{
	Use:     "job-schedule-shift-start-site-changes",
	Aliases: []string{"job-schedule-shift-start-site-change"},
	Short:   "Manage job schedule shift start site changes",
	Long:    "Commands for creating job schedule shift start site changes.",
}

func init() {
	doCmd.AddCommand(doJobScheduleShiftStartSiteChangesCmd)
}
