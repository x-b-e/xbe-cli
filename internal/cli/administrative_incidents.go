package cli

import "github.com/spf13/cobra"

var administrativeIncidentsCmd = &cobra.Command{
	Use:     "administrative-incidents",
	Aliases: []string{"administrative-incident"},
	Short:   "View administrative incidents",
	Long: `Browse administrative incidents.

Administrative incidents capture operational or administrative issues such as
capacity, planning, and quality concerns. Use these commands to list or
inspect administrative incident records.`,
	Example: `  # List administrative incidents
  xbe view administrative-incidents list

  # Show an administrative incident
  xbe view administrative-incidents show 123`,
}

func init() {
	viewCmd.AddCommand(administrativeIncidentsCmd)
}
