package cli

import "github.com/spf13/cobra"

var doIncidentUnitOfMeasureQuantitiesCmd = &cobra.Command{
	Use:     "incident-unit-of-measure-quantities",
	Aliases: []string{"incident-unit-of-measure-quantity"},
	Short:   "Manage incident unit of measure quantities",
	Long: `Create, update, and delete incident unit of measure quantities.

Incident unit of measure quantities store incident impact in specific units
(e.g., hours, tons, dollars).`,
	Example: `  # Create an incident unit of measure quantity
  xbe do incident-unit-of-measure-quantities create \
    --incident-type incidents --incident-id 123 \
    --unit-of-measure 456 --quantity 12.5

  # Update quantity
  xbe do incident-unit-of-measure-quantities update 789 --quantity 15

  # Delete a quantity
  xbe do incident-unit-of-measure-quantities delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doIncidentUnitOfMeasureQuantitiesCmd)
}
