package cli

import "github.com/spf13/cobra"

var hosViolationsCmd = &cobra.Command{
	Use:     "hos-violations",
	Aliases: []string{"hos-violation"},
	Short:   "Browse HOS violations",
	Long: `Browse HOS violations.

HOS violations capture hours-of-service rule breaches detected for a driver.

Commands:
  list    List HOS violations with filtering and pagination
  show    Show full details of an HOS violation`,
	Example: `  # List HOS violations
  xbe view hos-violations list

  # Filter by driver
  xbe view hos-violations list --driver 123

  # Filter by date range
  xbe view hos-violations list --start-at-min 2025-01-01T00:00:00Z --end-at-max 2025-01-02T00:00:00Z

  # Show violation details
  xbe view hos-violations show 456`,
}

func init() {
	viewCmd.AddCommand(hosViolationsCmd)
}
