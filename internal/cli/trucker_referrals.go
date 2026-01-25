package cli

import "github.com/spf13/cobra"

var truckerReferralsCmd = &cobra.Command{
	Use:     "trucker-referrals",
	Aliases: []string{"trucker-referral"},
	Short:   "View trucker referrals",
	Long:    "Commands for viewing trucker referrals.",
}

func init() {
	viewCmd.AddCommand(truckerReferralsCmd)
}
