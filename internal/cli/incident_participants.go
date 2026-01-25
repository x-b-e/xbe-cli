package cli

import "github.com/spf13/cobra"

var incidentParticipantsCmd = &cobra.Command{
	Use:     "incident-participants",
	Aliases: []string{"incident-participant"},
	Short:   "View incident participants",
	Long: `View participants associated with incidents.

Incident participants capture people involved in or impacted by an incident.

Commands:
  list    List incident participants
  show    Show incident participant details`,
	Example: `  # List incident participants
  xbe view incident-participants list

  # Filter by incident
  xbe view incident-participants list --incident 123

  # Show an incident participant
  xbe view incident-participants show 456`,
}

func init() {
	viewCmd.AddCommand(incidentParticipantsCmd)
}
