package cli

import "github.com/spf13/cobra"

var doProjectTransportPlanDriverConfirmationsCmd = &cobra.Command{
	Use:     "project-transport-plan-driver-confirmations",
	Aliases: []string{"project-transport-plan-driver-confirmation"},
	Short:   "Manage project transport plan driver confirmations",
	Long:    "Commands for updating project transport plan driver confirmations.",
}

func init() {
	doCmd.AddCommand(doProjectTransportPlanDriverConfirmationsCmd)
}
