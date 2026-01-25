package cli

import "github.com/spf13/cobra"

var doProjectTransportPlanStopInsertionsCmd = &cobra.Command{
	Use:     "project-transport-plan-stop-insertions",
	Aliases: []string{"project-transport-plan-stop-insertion"},
	Short:   "Manage project transport plan stop insertions",
	Long: `Create stop insertions that insert, move, or delete stops within a
project transport plan.`,
}

func init() {
	doCmd.AddCommand(doProjectTransportPlanStopInsertionsCmd)
}
