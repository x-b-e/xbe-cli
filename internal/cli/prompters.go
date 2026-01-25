package cli

import "github.com/spf13/cobra"

var promptersCmd = &cobra.Command{
	Use:   "prompters",
	Short: "Browse and view prompters",
	Long: `Browse and view prompters used for AI prompt templates.

Prompters store prompt templates that power release note suggestions, glossary
term definitions, safety risk summaries, and other AI-assisted workflows.

Commands:
  list    List prompters
  show    Show prompter details`,
	Example: `  # List prompters
  xbe view prompters list

  # Filter by active status
  xbe view prompters list --is-active true

  # Show a prompter
  xbe view prompters show 123`,
}

func init() {
	viewCmd.AddCommand(promptersCmd)
}
