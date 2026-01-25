package cli

import "github.com/spf13/cobra"

func newMaterialSitesShowCmd() *cobra.Command {
	return newGenericShowCmd("material-sites")
}

func init() {
	materialSitesCmd.AddCommand(newMaterialSitesShowCmd())
}
