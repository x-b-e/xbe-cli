package cli

import "github.com/spf13/cobra"

var equipmentMovementStopCompletionsCmd = &cobra.Command{
	Use:     "equipment-movement-stop-completions",
	Aliases: []string{"equipment-movement-stop-completion"},
	Short:   "View equipment movement stop completions",
	Long: `View equipment movement stop completions.

Equipment movement stop completions record when equipment movement stops
were completed, including optional location data and notes.

Commands:
  list  List equipment movement stop completions
  show  Show equipment movement stop completion details`,
}

func init() {
	viewCmd.AddCommand(equipmentMovementStopCompletionsCmd)
}
