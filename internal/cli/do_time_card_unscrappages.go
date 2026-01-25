package cli

import "github.com/spf13/cobra"

var doTimeCardUnscrappagesCmd = &cobra.Command{
	Use:     "time-card-unscrappages",
	Aliases: []string{"time-card-unscrappage"},
	Short:   "Manage time card unscrappages",
	Long: `Create time card unscrappages.

Commands:
  create    Create a time card unscrappage`,
}

func init() {
	doCmd.AddCommand(doTimeCardUnscrappagesCmd)
}
