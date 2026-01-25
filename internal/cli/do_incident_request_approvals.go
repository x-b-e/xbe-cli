package cli

import "github.com/spf13/cobra"

var doIncidentRequestApprovalsCmd = &cobra.Command{
	Use:     "incident-request-approvals",
	Aliases: []string{"incident-request-approval"},
	Short:   "Approve incident requests",
	Long: `Approve incident requests.

Incident request approvals transition submitted incident requests to approved and create incidents.

Commands:
  create    Approve an incident request`,
}

func init() {
	doCmd.AddCommand(doIncidentRequestApprovalsCmd)
}
