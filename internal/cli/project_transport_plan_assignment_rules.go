package cli

import "github.com/spf13/cobra"

var projectTransportPlanAssignmentRulesCmd = &cobra.Command{
	Use:   "project-transport-plan-assignment-rules",
	Short: "View project transport plan assignment rules",
	Long: `View project transport plan assignment rules.

Project transport plan assignment rules define broker-level rules for assigning
project transport plan drivers, tractors, and trailers.

Commands:
  list  List project transport plan assignment rules
  show  Show project transport plan assignment rule details`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanAssignmentRulesCmd)
}
