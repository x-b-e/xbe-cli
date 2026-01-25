package cli

import "github.com/spf13/cobra"

func newTrailersShowCmd() *cobra.Command {
	return newGenericShowCmd("trailers")
}

func init() {
	trailersCmd.AddCommand(newTrailersShowCmd())
}
