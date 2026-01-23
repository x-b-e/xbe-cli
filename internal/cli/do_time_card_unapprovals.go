package cli

import "github.com/spf13/cobra"

var doTimeCardUnapprovalsCmd = &cobra.Command{
	Use:     "time-card-unapprovals",
	Aliases: []string{"time-card-unapproval"},
	Short:   "Unapprove time cards",
	Long:    "Commands for unapproving time cards.",
}

func init() {
	doCmd.AddCommand(doTimeCardUnapprovalsCmd)
}
