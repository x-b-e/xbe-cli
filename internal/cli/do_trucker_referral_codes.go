package cli

import "github.com/spf13/cobra"

var doTruckerReferralCodesCmd = &cobra.Command{
	Use:     "trucker-referral-codes",
	Aliases: []string{"trucker-referral-code"},
	Short:   "Manage trucker referral codes",
	Long:    "Create, update, and delete trucker referral codes.",
}

func init() {
	doCmd.AddCommand(doTruckerReferralCodesCmd)
}
