package cli

import "github.com/spf13/cobra"

var doJobProductionPlanMaterialTypesCmd = &cobra.Command{
	Use:     "job-production-plan-material-types",
	Aliases: []string{"job-production-plan-material-type"},
	Short:   "Manage job production plan material types",
	Long:    "Create, update, and delete job production plan material types.",
}

func init() {
	doCmd.AddCommand(doJobProductionPlanMaterialTypesCmd)
}
