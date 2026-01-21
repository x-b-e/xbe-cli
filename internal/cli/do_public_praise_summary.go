package cli

import "github.com/spf13/cobra"

var doPublicPraiseSummaryCmd = &cobra.Command{
	Use:   "public-praise-summary",
	Short: "Generate public praise summaries",
	Long: `Generate public praise summaries on the XBE platform.

Commands:
  create    Generate a public praise summary`,
	Example: `  # Generate a public praise summary grouped by recipient
  xbe summarize public-praise-summary create --group-by recipient --filter broker=123 --filter created_at_min=2025-01-01 --filter created_at_max=2025-01-31

  # Summary by culture value
  xbe summarize public-praise-summary create --group-by culture_value --filter broker=123

  # Summary by giver and recipient
  xbe summarize public-praise-summary create --group-by given_by,recipient --filter broker=123`,
}

func init() {
	summarizeCmd.AddCommand(doPublicPraiseSummaryCmd)
}
