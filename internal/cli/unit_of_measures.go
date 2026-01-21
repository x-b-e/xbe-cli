package cli

import "github.com/spf13/cobra"

var unitOfMeasuresCmd = &cobra.Command{
	Use:   "unit-of-measures",
	Short: "View units of measure",
	Long: `View units of measure on the XBE platform.

Units of measure define how quantities are measured for billing and tracking
(e.g., tons, cubic yards, hours). They include information about the metric
type and measurement system.

Commands:
  list    List units of measure`,
	Example: `  # List units of measure
  xbe view unit-of-measures list

  # Filter by metric type
  xbe view unit-of-measures list --metric mass`,
}

func init() {
	viewCmd.AddCommand(unitOfMeasuresCmd)
}
