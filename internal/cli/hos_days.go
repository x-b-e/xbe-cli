package cli

import "github.com/spf13/cobra"

var hosDaysCmd = &cobra.Command{
	Use:     "hos-days",
	Aliases: []string{"hos-day"},
	Short:   "View HOS days",
	Long:    "Commands for viewing HOS days.",
}

func init() {
	viewCmd.AddCommand(hosDaysCmd)
}
