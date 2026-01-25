package cli

import "github.com/spf13/cobra"

var productionIncidentsCmd = &cobra.Command{
	Use:     "production-incidents",
	Aliases: []string{"production-incident"},
	Short:   "View production incidents",
	Long: `View production incidents.

Production incidents track production-impacting events like equipment delays,
trucking constraints, and job site conditions. Use the do commands to create
or update production incidents.`,
	Example: `  # List production incidents
  xbe view production-incidents list

  # Show a production incident
  xbe view production-incidents show 123`,
}

func init() {
	viewCmd.AddCommand(productionIncidentsCmd)
}
