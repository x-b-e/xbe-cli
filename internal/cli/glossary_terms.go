package cli

import "github.com/spf13/cobra"

var glossaryTermsCmd = &cobra.Command{
	Use:   "glossary-terms",
	Short: "Browse and view glossary terms",
	Long: `Browse and view glossary terms on the XBE platform.

Glossary terms provide definitions for industry terminology, product features,
and technical concepts used in heavy materials, logistics, and construction.

Commands:
  list    List glossary terms with filtering
  show    View the full details of a specific glossary term`,
	Example: `  # List all glossary terms
  xbe view glossary-terms list

  # Filter by source
  xbe view glossary-terms list --source xbe

  # View a specific glossary term
  xbe view glossary-terms show 123`,
}

func init() {
	viewCmd.AddCommand(glossaryTermsCmd)
}
