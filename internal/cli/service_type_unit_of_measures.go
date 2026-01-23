package cli

import "github.com/spf13/cobra"

var serviceTypeUnitOfMeasuresCmd = &cobra.Command{
	Use:     "service-type-unit-of-measures",
	Aliases: []string{"service-type-unit-of-measure"},
	Short:   "View service type unit of measures",
	Long: `View service type unit of measures on the XBE platform.

Service type unit of measures (STUOMs) combine a service type and unit of
measure to define how work is quantified and billed.

Commands:
  list    List service type unit of measures
  show    Show service type unit of measure details`,
	Example: `  # List service type unit of measures
  xbe view service-type-unit-of-measures list

  # View a specific service type unit of measure
  xbe view service-type-unit-of-measures show 123`,
}

func init() {
	viewCmd.AddCommand(serviceTypeUnitOfMeasuresCmd)
}
