package cli

import "github.com/spf13/cobra"

var doBaseSummaryTemplatesCmd = &cobra.Command{
	Use:     "base-summary-templates",
	Aliases: []string{"base-summary-template"},
	Short:   "Manage base summary templates",
	Long: `Manage base summary templates.

Commands:
  create    Create a base summary template
  delete    Delete a base summary template`,
	Example: `  # Create a base summary template
  xbe do base-summary-templates create --label "Weekly Summary" --broker 123

  # Delete a base summary template
  xbe do base-summary-templates delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doBaseSummaryTemplatesCmd)
}
