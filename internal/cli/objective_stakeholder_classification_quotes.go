package cli

import "github.com/spf13/cobra"

var objectiveStakeholderClassificationQuotesCmd = &cobra.Command{
	Use:     "objective-stakeholder-classification-quotes",
	Aliases: []string{"objective-stakeholder-classification-quote"},
	Short:   "Browse objective stakeholder classification quotes",
	Long: `Browse objective stakeholder classification quotes.

Objective stakeholder classification quotes capture narrative content tied
to objective stakeholder classifications and interest scoring.

Commands:
  list    List objective stakeholder classification quotes
  show    Show objective stakeholder classification quote details`,
	Example: `  # List quotes
  xbe view objective-stakeholder-classification-quotes list

  # Filter by classification
  xbe view objective-stakeholder-classification-quotes list --objective-stakeholder-classification 123

  # Filter by interest degree range
  xbe view objective-stakeholder-classification-quotes list --interest-degree-min 2 --interest-degree-max 5

  # Show a quote
  xbe view objective-stakeholder-classification-quotes show 456

  # JSON output
  xbe view objective-stakeholder-classification-quotes list --json`,
}

func init() {
	viewCmd.AddCommand(objectiveStakeholderClassificationQuotesCmd)
}
