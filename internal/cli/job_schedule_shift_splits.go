package cli

import "github.com/spf13/cobra"

var jobScheduleShiftSplitsCmd = &cobra.Command{
	Use:     "job-schedule-shift-splits",
	Aliases: []string{"job-schedule-shift-split"},
	Short:   "View job schedule shift splits",
	Long: `View job schedule shift splits.

Job schedule shift splits capture the creation of a new shift from an existing
flexible shift, optionally adjusting expected loads or start times.

Commands:
  list    List job schedule shift splits
  show    Show job schedule shift split details`,
	Example: `  # List shift splits
  xbe view job-schedule-shift-splits list

  # Show a shift split
  xbe view job-schedule-shift-splits show 123`,
}

func init() {
	viewCmd.AddCommand(jobScheduleShiftSplitsCmd)
}
