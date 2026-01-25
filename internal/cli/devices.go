package cli

import "github.com/spf13/cobra"

var devicesCmd = &cobra.Command{
	Use:     "devices",
	Aliases: []string{"device"},
	Short:   "View devices",
	Long:    "Commands for viewing devices (mobile app instances).",
}

func init() {
	viewCmd.AddCommand(devicesCmd)
}
