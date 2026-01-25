package cli

import "github.com/spf13/cobra"

var doPromptersCmd = &cobra.Command{
	Use:   "prompters",
	Short: "Manage prompters",
	Long: `Manage prompters on the XBE platform.

Prompters store prompt templates used for AI-assisted workflows such as release
note summaries, glossary term definitions, and safety risk suggestions.

Commands:
  create    Create a new prompter
  update    Update an existing prompter
  delete    Delete a prompter`,
	Example: `  # Create a prompter
  xbe do prompters create --name "Release Notes" --is-active=false

  # Update a prompt template
  xbe do prompters update 123 --release-note-headline-suggestions-prompt-template "New template"

  # Delete a prompter (requires --confirm)
  xbe do prompters delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doPromptersCmd)
}
