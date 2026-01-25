package cli

import "github.com/spf13/cobra"

var incidentHeadlineSuggestionsCmd = &cobra.Command{
	Use:     "incident-headline-suggestions",
	Aliases: []string{"incident-headline-suggestion"},
	Short:   "Browse incident headline suggestions",
	Long: `Browse incident headline suggestions.

Headline suggestions provide AI-generated, single-sentence summaries
of incidents for use in reports and dashboards.

Commands:
  list    List incident headline suggestions
  show    Show incident headline suggestion details`,
	Example: `  # List incident headline suggestions
  xbe view incident-headline-suggestions list

  # Filter by incident
  xbe view incident-headline-suggestions list --incident 123

  # Show a suggestion
  xbe view incident-headline-suggestions show 456`,
}

func init() {
	viewCmd.AddCommand(incidentHeadlineSuggestionsCmd)
}
