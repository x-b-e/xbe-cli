package cli

import "github.com/spf13/cobra"

var doIncidentRequestRejectionsCmd = &cobra.Command{
	Use:     "incident-request-rejections",
	Aliases: []string{"incident-request-rejection"},
	Short:   "Reject incident requests",
	Long: `Reject incident requests.

Incident request rejections transition submitted incident requests to rejected.

Commands:
  create    Reject an incident request`,
}

func init() {
	doCmd.AddCommand(doIncidentRequestRejectionsCmd)
}
