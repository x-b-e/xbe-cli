package cli

import "github.com/spf13/cobra"

var materialTransactionSubmissionsCmd = &cobra.Command{
	Use:     "material-transaction-submissions",
	Aliases: []string{"material-transaction-submission"},
	Short:   "View material transaction submissions",
	Long: `View material transaction submissions.

Submissions record a status change to submitted for a material transaction and may
include a comment.

Commands:
  list    List material transaction submissions
  show    Show material transaction submission details`,
	Example: `  # List submissions
  xbe view material-transaction-submissions list

  # Show a submission
  xbe view material-transaction-submissions show 123`,
}

func init() {
	viewCmd.AddCommand(materialTransactionSubmissionsCmd)
}
