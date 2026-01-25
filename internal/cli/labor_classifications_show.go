package cli

import "github.com/spf13/cobra"

func newLaborClassificationsShowCmd() *cobra.Command {
	return newGenericShowCmd("labor-classifications")
}

func init() {
	laborClassificationsCmd.AddCommand(newLaborClassificationsShowCmd())
}
