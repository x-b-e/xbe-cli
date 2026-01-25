package cli

import "github.com/spf13/cobra"

var efficiencyIncidentsCmd = &cobra.Command{
	Use:     "efficiency-incidents",
	Aliases: []string{"efficiency-incident"},
	Short:   "View efficiency incidents",
	Long: `Browse efficiency incidents.

Efficiency incidents capture operational inefficiencies such as over trucking.
Use these commands to list or inspect efficiency incident records.`,
	Example: `  # List efficiency incidents
  xbe view efficiency-incidents list

  # Show an efficiency incident
  xbe view efficiency-incidents show 123`,
}

func init() {
	viewCmd.AddCommand(efficiencyIncidentsCmd)
}
