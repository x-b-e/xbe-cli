package cli

import "github.com/spf13/cobra"

var serviceTypeUnitOfMeasureQuantitiesCmd = &cobra.Command{
	Use:   "service-type-unit-of-measure-quantities",
	Short: "View service type unit of measure quantities",
	Long: `View service type unit of measure quantities on the XBE platform.

Service type unit of measure quantities represent quantified amounts on
resources such as time cards, including calculated and explicit quantities.

Commands:
  list    List service type unit of measure quantities
  show    Show service type unit of measure quantity details`,
	Example: `  # List service type unit of measure quantities
  xbe view service-type-unit-of-measure-quantities list

  # Filter by service type unit of measure
  xbe view service-type-unit-of-measure-quantities list --service-type-unit-of-measure 123

  # Show a specific quantity
  xbe view service-type-unit-of-measure-quantities show 456`,
}

func init() {
	viewCmd.AddCommand(serviceTypeUnitOfMeasureQuantitiesCmd)
}
