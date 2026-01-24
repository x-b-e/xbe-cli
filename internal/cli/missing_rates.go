package cli

import "github.com/spf13/cobra"

var missingRatesCmd = &cobra.Command{
	Use:     "missing-rates",
	Aliases: []string{"missing-rate"},
	Short:   "View missing rates",
	Long: `Browse missing rates.

Missing rates are created when a job is missing a rate for a service type
unit of measure. Creating a missing rate adds rates to customer and broker
side tenders tied to the job.

Commands:
  list    List missing rates with filtering and pagination
  show    View the full details of a missing rate`,
}

func init() {
	viewCmd.AddCommand(missingRatesCmd)
}
