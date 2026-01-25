package cli

import "github.com/spf13/cobra"

var doIncidentHeadlineSuggestionsCmd = &cobra.Command{
	Use:     "incident-headline-suggestions",
	Aliases: []string{"incident-headline-suggestion"},
	Short:   "Manage incident headline suggestions",
	Long: `Create AI-generated headline suggestions for incidents.

Commands:
  create    Create an incident headline suggestion
  delete    Delete an incident headline suggestion`,
	Example: `  # Create a headline suggestion
  xbe do incident-headline-suggestions create --incident 123

  # Create with custom options
  xbe do incident-headline-suggestions create --incident 123 --options '{"temperature":0.5,"max_tokens":256}'

  # Output as JSON
  xbe do incident-headline-suggestions create --incident 123 --json

  # Delete a suggestion
  xbe do incident-headline-suggestions delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doIncidentHeadlineSuggestionsCmd)
}
