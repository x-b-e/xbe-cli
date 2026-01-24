package cli

import "github.com/spf13/cobra"

var doObjectiveStakeholderClassificationQuotesCmd = &cobra.Command{
	Use:   "objective-stakeholder-classification-quotes",
	Short: "Manage objective stakeholder classification quotes",
	Long: `Create, update, and delete objective stakeholder classification quotes.

Commands:
  create    Create an objective stakeholder classification quote
  update    Update an objective stakeholder classification quote
  delete    Delete an objective stakeholder classification quote`,
	Example: `  # Create a quote
  xbe do objective-stakeholder-classification-quotes create \
    --objective-stakeholder-classification 123 \
    --content "Stakeholder values transparency" \
    --is-generated

  # Update a quote
  xbe do objective-stakeholder-classification-quotes update 456 --content "Updated content"

  # Delete a quote
  xbe do objective-stakeholder-classification-quotes delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doObjectiveStakeholderClassificationQuotesCmd)
}
