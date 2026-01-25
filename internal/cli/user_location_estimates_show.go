package cli

import "github.com/spf13/cobra"

func newUserLocationEstimatesShowCmd() *cobra.Command {
	return newGenericShowCmd("user-location-estimates")
}

func init() {
	userLocationEstimatesCmd.AddCommand(newUserLocationEstimatesShowCmd())
}
