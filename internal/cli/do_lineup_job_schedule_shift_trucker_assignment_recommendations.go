package cli

import "github.com/spf13/cobra"

var doLineupJobScheduleShiftTruckerAssignmentRecommendationsCmd = &cobra.Command{
	Use:     "lineup-job-schedule-shift-trucker-assignment-recommendations",
	Aliases: []string{"lineup-job-schedule-shift-trucker-assignment-recommendation"},
	Short:   "Manage lineup job schedule shift trucker assignment recommendations",
	Long: `Create trucker assignment recommendations for lineup job schedule shifts.

Commands:
  create    Generate recommendations for a shift`,
	Example: `  # Generate recommendations for a lineup job schedule shift
  xbe do lineup-job-schedule-shift-trucker-assignment-recommendations create --lineup-job-schedule-shift 123`,
}

func init() {
	doCmd.AddCommand(doLineupJobScheduleShiftTruckerAssignmentRecommendationsCmd)
}
