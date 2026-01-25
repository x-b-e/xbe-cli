package cli

import "github.com/spf13/cobra"

func newPlacePredictionsShowCmd() *cobra.Command {
	return newGenericShowCmd("place-predictions")
}

func init() {
	placePredictionsCmd.AddCommand(newPlacePredictionsShowCmd())
}
