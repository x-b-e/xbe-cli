package cli

import "github.com/spf13/cobra"

var doTenderRejectionsCmd = &cobra.Command{
	Use:     "tender-rejections",
	Aliases: []string{"tender-rejection"},
	Short:   "Reject tenders",
	Long:    "Commands for rejecting tenders.",
}

func init() {
	doCmd.AddCommand(doTenderRejectionsCmd)
}
