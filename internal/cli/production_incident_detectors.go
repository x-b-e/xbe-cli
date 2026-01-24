package cli

import "github.com/spf13/cobra"

var productionIncidentDetectorsCmd = &cobra.Command{
	Use:   "production-incident-detectors",
	Short: "Browse production incident detector runs",
	Long: `Browse production incident detector runs on the XBE platform.

Production incident detectors analyze job production plans to identify
periods of under- or over-production based on configurable thresholds.

Commands:
  list    List detector runs
  show    Show detector run details

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
	Example: `  # List detector runs
  xbe view production-incident-detectors list

  # Show detector results
  xbe view production-incident-detectors show 123`,
}

func init() {
	viewCmd.AddCommand(productionIncidentDetectorsCmd)
}
