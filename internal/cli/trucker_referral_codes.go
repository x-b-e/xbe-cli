package cli

import "github.com/spf13/cobra"

var truckerReferralCodesCmd = &cobra.Command{
	Use:     "trucker-referral-codes",
	Aliases: []string{"trucker-referral-code"},
	Short:   "Browse trucker referral codes",
	Long:    "Browse trucker referral codes.",
}

func init() {
	viewCmd.AddCommand(truckerReferralCodesCmd)
}
