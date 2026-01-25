package cli

import "github.com/spf13/cobra"

var doJobProductionPlanTrailerClassificationsCmd = &cobra.Command{
	Use:     "job-production-plan-trailer-classifications",
	Aliases: []string{"job-production-plan-trailer-classification"},
	Short:   "Manage job production plan trailer classifications",
	Long: `Manage job production plan trailer classifications.

Job production plan trailer classifications define which trailer
classifications apply to a job production plan, including optional
weight and material transaction limits.

Commands:
  create  Add a trailer classification to a job production plan
  update  Update trailer classification settings or relationships
  delete  Remove a trailer classification from a job production plan`,
	Example: `  # Add a trailer classification to a job production plan
  xbe do job-production-plan-trailer-classifications create \
    --job-production-plan 123 \
    --trailer-classification 456 \
    --gross-weight-legal-limit-lbs-explicit 80000 \
    --explicit-material-transaction-tons-max 20

  # Update equivalent trailer classifications
  xbe do job-production-plan-trailer-classifications update 789 \
    --trailer-classification-equivalent-ids 111,222

  # Delete a job production plan trailer classification
  xbe do job-production-plan-trailer-classifications delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanTrailerClassificationsCmd)
}
