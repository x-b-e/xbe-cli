package cli

import "github.com/spf13/cobra"

var doTenderRatersCmd = &cobra.Command{
	Use:     "tender-raters",
	Aliases: []string{"tender-rater"},
	Short:   "Rate tenders against applicable rate agreements",
	Long: `Rate tenders against applicable rate agreements.

Tender raters calculate rate agreement rates and shift set time card
constraints for a tender. By default, results are computed without
persisting changes. Use --persist-changes true to write updates back to
the tender.`,
	Example: `  # Rate a tender without persisting changes
  xbe do tender-raters create --tender 123

  # Rate a tender and persist changes
  xbe do tender-raters create --tender 123 --persist-changes true`,
}

func init() {
	doCmd.AddCommand(doTenderRatersCmd)
}
