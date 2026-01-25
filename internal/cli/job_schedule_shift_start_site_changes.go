package cli

import "github.com/spf13/cobra"

var jobScheduleShiftStartSiteChangesCmd = &cobra.Command{
	Use:     "job-schedule-shift-start-site-changes",
	Aliases: []string{"job-schedule-shift-start-site-change"},
	Short:   "View job schedule shift start site changes",
	Long: `View job schedule shift start site changes.

Job schedule shift start site changes capture updates to a shift's start site
and record the previous and new start sites.

Commands:
  list    List job schedule shift start site changes
  show    Show job schedule shift start site change details`,
}

func init() {
	viewCmd.AddCommand(jobScheduleShiftStartSiteChangesCmd)
}
