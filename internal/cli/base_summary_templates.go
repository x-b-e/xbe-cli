package cli

import "github.com/spf13/cobra"

var baseSummaryTemplatesCmd = &cobra.Command{
	Use:     "base-summary-templates",
	Aliases: []string{"base-summary-template"},
	Short:   "Browse base summary templates",
	Long: `Browse base summary templates.

Base summary templates define reusable summary configurations such as groupings,
filters, and explicit metrics for reporting views.`,
	Example: `  # List base summary templates
  xbe view base-summary-templates list

  # Show a base summary template
  xbe view base-summary-templates show 123`,
}

func init() {
	viewCmd.AddCommand(baseSummaryTemplatesCmd)
}
