package cli

import "github.com/spf13/cobra"

var brokersCmd = &cobra.Command{
	Use:   "brokers",
	Short: "View brokers (branches)",
	Long:  "View brokers (also called branches).",
}

func init() {
	viewCmd.AddCommand(brokersCmd)
}
