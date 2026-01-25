package cli

import "github.com/spf13/cobra"

var doRateAgreementCopierWorksCmd = &cobra.Command{
	Use:     "rate-agreement-copier-works",
	Aliases: []string{"rate-agreement-copier-work"},
	Short:   "Manage rate agreement copier works",
	Long: `Commands for creating and updating rate agreement copier works.

Use these commands to kick off rate agreement copy jobs and adjust metadata
such as notes.`,
}

func init() {
	doCmd.AddCommand(doRateAgreementCopierWorksCmd)
}
