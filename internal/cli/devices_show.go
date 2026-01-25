package cli

import "github.com/spf13/cobra"

func newDevicesShowCmd() *cobra.Command {
	return newGenericShowCmd("devices")
}

func init() {
	devicesCmd.AddCommand(newDevicesShowCmd())
}
