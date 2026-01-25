package cli

import "github.com/spf13/cobra"

var doDigitalFleetTrucksCmd = &cobra.Command{
	Use:     "digital-fleet-trucks",
	Aliases: []string{"digital-fleet-truck"},
	Short:   "Manage digital fleet trucks",
	Long: `Manage Digital Fleet truck assignments on the XBE platform.

Digital fleet trucks are created from integrations. They cannot be created or
removed via the API, but trailer and tractor assignments can be updated.`,
}

func init() {
	doCmd.AddCommand(doDigitalFleetTrucksCmd)
}
