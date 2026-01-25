package cli

import "github.com/spf13/cobra"

var doProjectTransportPlanAssignmentRulesCmd = &cobra.Command{
	Use:     "project-transport-plan-assignment-rules",
	Aliases: []string{"project-transport-plan-assignment-rule"},
	Short:   "Manage project transport plan assignment rules",
	Long:    "Create, update, and delete project transport plan assignment rules.",
}

func init() {
	doCmd.AddCommand(doProjectTransportPlanAssignmentRulesCmd)
}
