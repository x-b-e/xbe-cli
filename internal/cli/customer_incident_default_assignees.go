package cli

import "github.com/spf13/cobra"

var customerIncidentDefaultAssigneesCmd = &cobra.Command{
	Use:     "customer-incident-default-assignees",
	Aliases: []string{"customer-incident-default-assignee"},
	Short:   "View customer incident default assignees",
	Long:    "Commands for viewing customer incident default assignees.",
}

func init() {
	viewCmd.AddCommand(customerIncidentDefaultAssigneesCmd)
}
