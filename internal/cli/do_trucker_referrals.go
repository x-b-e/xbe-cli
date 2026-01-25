package cli

import "github.com/spf13/cobra"

var doTruckerReferralsCmd = &cobra.Command{
	Use:     "trucker-referrals",
	Aliases: []string{"trucker-referral"},
	Short:   "Manage trucker referrals",
	Long:    "Commands for creating, updating, and deleting trucker referrals.",
}

func init() {
	doCmd.AddCommand(doTruckerReferralsCmd)
}
