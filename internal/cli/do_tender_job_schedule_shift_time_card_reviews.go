package cli

import "github.com/spf13/cobra"

var doTenderJobScheduleShiftTimeCardReviewsCmd = &cobra.Command{
	Use:     "tender-job-schedule-shift-time-card-reviews",
	Aliases: []string{"tender-job-schedule-shift-time-card-review"},
	Short:   "Manage tender job schedule shift time card reviews",
	Long: `Create and delete tender job schedule shift time card reviews.

Commands:
  create    Create a time card review
  delete    Delete a time card review`,
}

func init() {
	doCmd.AddCommand(doTenderJobScheduleShiftTimeCardReviewsCmd)
}
