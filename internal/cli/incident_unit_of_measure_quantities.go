package cli

import "github.com/spf13/cobra"

var incidentUnitOfMeasureQuantitiesCmd = &cobra.Command{
	Use:     "incident-unit-of-measure-quantities",
	Aliases: []string{"incident-unit-of-measure-quantity"},
	Short:   "View incident unit of measure quantities",
	Long: `View incident unit of measure quantities.

Incident unit of measure quantities track incident impact in specific units
(e.g., hours, tons, dollars) for reporting and analysis.

Commands:
  list    List incident unit of measure quantities
  show    Show incident unit of measure quantity details`,
	Example: `  # List incident unit of measure quantities
  xbe view incident-unit-of-measure-quantities list

  # Filter by incident
  xbe view incident-unit-of-measure-quantities list --incident-type incidents --incident-id 123

  # Show a specific quantity
  xbe view incident-unit-of-measure-quantities show 456

  # Output JSON
  xbe view incident-unit-of-measure-quantities list --json`,
}

func init() {
	viewCmd.AddCommand(incidentUnitOfMeasureQuantitiesCmd)
}
