package cli

import "github.com/spf13/cobra"

var doRmtAdjustmentsCmd = &cobra.Command{
	Use:     "rmt-adjustments",
	Aliases: []string{"rmt-adjustment"},
	Short:   "Adjust raw material transactions",
	Long: `Adjust raw material transactions.

This operation applies adjustments to raw material transactions (RMTs) and
records the adjustment details for auditing.

Commands:
  create    Create an RMT adjustment`,
	Example: `  # Adjust RMTs
  xbe do rmt-adjustments create --rmt-ids 123,456 --note "Corrected weights" \
    --raw-data-adjustments '{"net_weight":12.5,"net_weight_by_xbe_reason":"Scale correction"}'

  # Adjust an invoiced RMT
  xbe do rmt-adjustments create --rmt-ids 123 --note "Voided ticket" \
    --raw-data-adjustments '{"is_voided":true,"is_voided_by_xbe_reason":"Duplicate"}' \
    --update-if-invoiced`,
}

func init() {
	doCmd.AddCommand(doRmtAdjustmentsCmd)
}
