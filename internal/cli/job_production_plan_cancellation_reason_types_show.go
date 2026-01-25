package cli

import "github.com/spf13/cobra"

func newJobProductionPlanCancellationReasonTypesShowCmd() *cobra.Command {
	return newGenericShowCmd("job-production-plan-cancellation-reason-types")
}

func init() {
	jobProductionPlanCancellationReasonTypesCmd.AddCommand(newJobProductionPlanCancellationReasonTypesShowCmd())
}
