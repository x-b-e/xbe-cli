package cli

import "github.com/spf13/cobra"

var hosAvailabilitySnapshotsCmd = &cobra.Command{
	Use:     "hos-availability-snapshots",
	Aliases: []string{"hos-availability-snapshot"},
	Short:   "Browse HOS availability snapshots",
	Long: `Browse HOS availability snapshots.

HOS availability snapshots capture a driver's remaining hours-of-service
availability at a point in time.

Commands:
  list    List availability snapshots with filtering and pagination
  show    Show full details of an availability snapshot`,
	Example: `  # List snapshots
  xbe view hos-availability-snapshots list

  # Filter by driver
  xbe view hos-availability-snapshots list --driver 123

  # Filter by HOS day
  xbe view hos-availability-snapshots list --hos-day 456

  # Show snapshot details
  xbe view hos-availability-snapshots show 789`,
}

func init() {
	viewCmd.AddCommand(hosAvailabilitySnapshotsCmd)
}
