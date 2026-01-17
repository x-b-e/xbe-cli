package cli

import "github.com/spf13/cobra"

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View XBE content",
}

func init() {
	rootCmd.AddCommand(viewCmd)
}
