package cli

import "github.com/spf13/cobra"

var liabilityIncidentsCmd = &cobra.Command{
	Use:     "liability-incidents",
	Aliases: []string{"liability-incident"},
	Short:   "View liability incidents",
	Long: `View liability incidents.

Liability incidents track damage, theft, vandalism, and other liability events.
Use the do commands to create or update liability incidents.`,
	Example: `  # List liability incidents
  xbe view liability-incidents list

  # Show a liability incident
  xbe view liability-incidents show 123`,
}

func init() {
	viewCmd.AddCommand(liabilityIncidentsCmd)
}
