package cli

import "github.com/spf13/cobra"

var doMaintenanceRequirementRuleEvaluationClerksCmd = &cobra.Command{
	Use:     "maintenance-requirement-rule-evaluation-clerks",
	Aliases: []string{"maintenance-requirement-rule-evaluation-clerk"},
	Short:   "Manage maintenance requirement rule evaluation clerks",
	Long: `Trigger maintenance requirement rule evaluations for equipment.

Commands:
  create    Request evaluation for a piece of equipment`,
	Example: `  # Evaluate maintenance requirement rules for equipment
  xbe do maintenance-requirement-rule-evaluation-clerks create --equipment 123`,
}

func init() {
	doCmd.AddCommand(doMaintenanceRequirementRuleEvaluationClerksCmd)
}
