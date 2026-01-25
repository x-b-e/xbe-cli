package cli

import "github.com/spf13/cobra"

var doCommitmentItemsCmd = &cobra.Command{
	Use:     "commitment-items",
	Aliases: []string{"commitment-item"},
	Short:   "Manage commitment items",
	Long: `Create, update, and delete commitment items.

Commitment items define scheduling and adjustment rules for commitments.`,
}

func init() {
	doCmd.AddCommand(doCommitmentItemsCmd)
}
