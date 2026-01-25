package cli

import "github.com/spf13/cobra"

var doRateAgreementsCopiersCmd = &cobra.Command{
	Use:     "rate-agreements-copiers",
	Aliases: []string{"rate-agreements-copier"},
	Short:   "Copy rate agreements",
	Long:    "Commands for copying rate agreements to customers or truckers.",
}

func init() {
	doCmd.AddCommand(doRateAgreementsCopiersCmd)
}
