package cli

import "github.com/spf13/cobra"

var doMaterialSiteReadingSummaryCmd = &cobra.Command{
	Use:   "material-site-reading-summary",
	Short: "Generate material site reading summaries",
	Long: `Generate material site reading summaries on the XBE platform.

Commands:
  create    Generate a material site reading summary`,
	Example: `  # Minute-level summary for a material site measure
  xbe summarize material-site-reading-summary create --group-by minute \
    --filter material_site=123 --filter material_site_measure=456 \
    --filter reading_at_min=2025-01-01T00:00:00Z --filter reading_at_max=2025-01-01T00:30:00Z

  # Hourly summary with material type presence filter
  xbe summarize material-site-reading-summary create --group-by hour \
    --filter material_site=123 --filter material_site_measure=456 \
    --filter reading_at_min=2025-01-01T00:00:00Z --filter reading_at_max=2025-01-01T12:00:00Z \
    --filter material_site_reading_material_type_presence=true`,
}

func init() {
	summarizeCmd.AddCommand(doMaterialSiteReadingSummaryCmd)
}
