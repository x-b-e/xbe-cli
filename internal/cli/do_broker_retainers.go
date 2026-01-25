package cli

import "github.com/spf13/cobra"

var doBrokerRetainersCmd = &cobra.Command{
	Use:     "broker-retainers",
	Aliases: []string{"broker-retainer"},
	Short:   "Manage broker retainers",
	Long:    "Commands for creating, updating, and deleting broker retainers.",
}

func init() {
	doCmd.AddCommand(doBrokerRetainersCmd)
}
