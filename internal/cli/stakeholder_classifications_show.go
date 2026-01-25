package cli

import "github.com/spf13/cobra"

func newStakeholderClassificationsShowCmd() *cobra.Command {
	return newGenericShowCmd("stakeholder-classifications")
}

func init() {
	stakeholderClassificationsCmd.AddCommand(newStakeholderClassificationsShowCmd())
}
