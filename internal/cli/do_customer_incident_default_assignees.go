package cli

import "github.com/spf13/cobra"

var doCustomerIncidentDefaultAssigneesCmd = &cobra.Command{
	Use:     "customer-incident-default-assignees",
	Aliases: []string{"customer-incident-default-assignee"},
	Short:   "Manage customer incident default assignees",
	Long:    "Create, update, and delete customer incident default assignees.",
}

func init() {
	doCmd.AddCommand(doCustomerIncidentDefaultAssigneesCmd)
}
