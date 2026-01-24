package cli

import "github.com/spf13/cobra"

var tenderJobScheduleShiftTimeCardReviewsCmd = &cobra.Command{
	Use:     "tender-job-schedule-shift-time-card-reviews",
	Aliases: []string{"tender-job-schedule-shift-time-card-review"},
	Short:   "View tender job schedule shift time card reviews",
	Long: `View tender job schedule shift time card reviews.

Time card reviews capture automated analysis of shift time cards and
suggested start/end times and down minutes.

Commands:
  list    List time card reviews
  show    Show time card review details`,
	Example: `  # List time card reviews
  xbe view tender-job-schedule-shift-time-card-reviews list

  # Show a time card review
  xbe view tender-job-schedule-shift-time-card-reviews show 123

  # Output JSON
  xbe view tender-job-schedule-shift-time-card-reviews list --json`,
}

func init() {
	viewCmd.AddCommand(tenderJobScheduleShiftTimeCardReviewsCmd)
}
