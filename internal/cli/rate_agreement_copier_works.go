package cli

import "github.com/spf13/cobra"

var rateAgreementCopierWorksCmd = &cobra.Command{
	Use:     "rate-agreement-copier-works",
	Aliases: []string{"rate-agreement-copier-work"},
	Short:   "Browse rate agreement copier works",
	Long: `Browse rate agreement copier works.

Rate agreement copier works track background jobs that copy a rate agreement
between organizations.

Commands:
  list    List copier works with filtering and pagination
  show    Show full details of a copier work item`,
	Example: `  # List copier works
  xbe view rate-agreement-copier-works list

  # Show copier work details
  xbe view rate-agreement-copier-works show 123`,
}

func init() {
	viewCmd.AddCommand(rateAgreementCopierWorksCmd)
}
