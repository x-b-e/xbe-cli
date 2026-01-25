package cli

import "github.com/spf13/cobra"

func newIncidentsShowCmd() *cobra.Command {
	return newGenericShowCmd("incidents")
}

func init() {
	incidentsCmd.AddCommand(newIncidentsShowCmd())
}
