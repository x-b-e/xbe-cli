package cli

import "github.com/spf13/cobra"

var deviceDiagnosticsCmd = &cobra.Command{
	Use:     "device-diagnostics",
	Aliases: []string{"device-diagnostic"},
	Short:   "View device diagnostics",
	Long: `View device diagnostics collected from mobile devices.

Device diagnostics capture tracking state, permissions, and device environment
snapshots to help troubleshoot mobile tracking issues.

Commands:
  list    List device diagnostics
  show    Show device diagnostic details`,
	Example: `  # List recent device diagnostics
  xbe view device-diagnostics list --limit 10

  # Filter by device identifier
  xbe view device-diagnostics list --device-identifier "ABC-123"

  # Show a device diagnostic
  xbe view device-diagnostics show 456`,
}

func init() {
	viewCmd.AddCommand(deviceDiagnosticsCmd)
}
