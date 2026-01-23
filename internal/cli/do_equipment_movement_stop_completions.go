package cli

import "github.com/spf13/cobra"

var doEquipmentMovementStopCompletionsCmd = &cobra.Command{
	Use:     "equipment-movement-stop-completions",
	Aliases: []string{"equipment-movement-stop-completion"},
	Short:   "Manage equipment movement stop completions",
	Long:    "Commands for creating, updating, and deleting equipment movement stop completions.",
}

func init() {
	doCmd.AddCommand(doEquipmentMovementStopCompletionsCmd)
}
