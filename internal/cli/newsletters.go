package cli

import "github.com/spf13/cobra"

var newslettersCmd = &cobra.Command{
	Use:   "newsletters",
	Short: "View newsletters",
}

func init() {
	viewCmd.AddCommand(newslettersCmd)
}
