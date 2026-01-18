package cli

import "github.com/spf13/cobra"

var doGlossaryTermsCmd = &cobra.Command{
	Use:   "glossary-terms",
	Short: "Manage glossary terms",
	Long: `Manage glossary terms on the XBE platform.

Commands:
  create    Create a new glossary term
  update    Update an existing glossary term
  delete    Delete a glossary term`,
	Example: `  # Create a glossary term
  xbe do glossary-terms create --term "Paving" --definition "The process of laying asphalt"

  # Update a glossary term's definition
  xbe do glossary-terms update 123 --definition "New definition"

  # Delete a glossary term (requires --confirm)
  xbe do glossary-terms delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doGlossaryTermsCmd)
}
