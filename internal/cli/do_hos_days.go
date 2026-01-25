package cli

import "github.com/spf13/cobra"

var doHosDaysCmd = &cobra.Command{
	Use:     "hos-days",
	Aliases: []string{"hos-day"},
	Short:   "Manage HOS days",
	Long:    "Commands for updating HOS days.",
}

func init() {
	doCmd.AddCommand(doHosDaysCmd)
}
