package cli

import "github.com/spf13/cobra"

var doDevicesCmd = &cobra.Command{
	Use:     "devices",
	Aliases: []string{"device"},
	Short:   "Manage devices",
	Long:    "Commands for updating devices. Note: Devices cannot be created or deleted via the API.",
}

func init() {
	doCmd.AddCommand(doDevicesCmd)
}
