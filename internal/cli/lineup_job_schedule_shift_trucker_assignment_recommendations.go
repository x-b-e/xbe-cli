package cli

import "github.com/spf13/cobra"

var lineupJobScheduleShiftTruckerAssignmentRecommendationsCmd = &cobra.Command{
	Use:     "lineup-job-schedule-shift-trucker-assignment-recommendations",
	Aliases: []string{"lineup-job-schedule-shift-trucker-assignment-recommendation"},
	Short:   "Browse lineup job schedule shift trucker assignment recommendations",
	Long: `Browse lineup job schedule shift trucker assignment recommendations.

Recommendations rank truckers for a specific lineup job schedule shift.

Commands:
  list    List recommendations with filtering and pagination
  show    Show full recommendation details`,
	Example: `  # List recommendations
  xbe view lineup-job-schedule-shift-trucker-assignment-recommendations list

  # Filter by lineup job schedule shift
  xbe view lineup-job-schedule-shift-trucker-assignment-recommendations list --lineup-job-schedule-shift 123

  # Show a recommendation
  xbe view lineup-job-schedule-shift-trucker-assignment-recommendations show 456`,
}

func init() {
	viewCmd.AddCommand(lineupJobScheduleShiftTruckerAssignmentRecommendationsCmd)
}
