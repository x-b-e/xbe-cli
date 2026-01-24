package cli

import "github.com/spf13/cobra"

var doRateAdjustmentsCmd = &cobra.Command{
	Use:   "rate-adjustments",
	Short: "Manage rate adjustments",
	Long: `Create, update, and delete rate adjustments.

Rate adjustments tie rates to cost indexes and define how pricing is adjusted.

Commands:
  create  Create a rate adjustment
  update  Update a rate adjustment
  delete  Delete a rate adjustment`,
	Example: `  # Create a rate adjustment
  xbe do rate-adjustments create --rate 123 --cost-index 456 \
    --zero-intercept-value 100 --zero-intercept-ratio 0.25

  # Update a rate adjustment
  xbe do rate-adjustments update 123 --adjustment-max 5.0

  # Delete a rate adjustment
  xbe do rate-adjustments delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doRateAdjustmentsCmd)
}
