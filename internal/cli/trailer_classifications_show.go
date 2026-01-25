package cli

import "github.com/spf13/cobra"

func newTrailerClassificationsShowCmd() *cobra.Command {
	return newGenericShowCmd("trailer-classifications")
}

func init() {
	trailerClassificationsCmd.AddCommand(newTrailerClassificationsShowCmd())
}
