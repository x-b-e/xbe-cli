package cli

import "github.com/spf13/cobra"

var doDeviceDiagnosticsCmd = &cobra.Command{
	Use:     "device-diagnostics",
	Aliases: []string{"device-diagnostic"},
	Short:   "Manage device diagnostics",
	Long: `Create device diagnostic snapshots on the XBE platform.

Device diagnostics are typically created by mobile devices to report
tracking state, permissions, and device health at a point in time.

Commands:
  create  Create a device diagnostic`,
	Example: `  # Create a device diagnostic for a device identifier
  xbe do device-diagnostics create --device-identifier "ABC-123" --is-tracking=true`,
}

func init() {
	doCmd.AddCommand(doDeviceDiagnosticsCmd)
}
