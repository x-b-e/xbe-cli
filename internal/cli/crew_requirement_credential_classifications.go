package cli

import "github.com/spf13/cobra"

var crewRequirementCredentialClassificationsCmd = &cobra.Command{
	Use:     "crew-requirement-credential-classifications",
	Aliases: []string{"crew-requirement-credential-classification"},
	Short:   "Browse crew requirement credential classifications",
	Long: `Browse crew requirement credential classifications.

Crew requirement credential classifications link crew requirements to the
credential classifications they require.

Commands:
  list    List links with filtering and pagination
  show    Show full details of a link`,
	Example: `  # List crew requirement credential classifications
  xbe view crew-requirement-credential-classifications list

  # Filter by crew requirement
  xbe view crew-requirement-credential-classifications list --crew-requirement 123

  # Show a link
  xbe view crew-requirement-credential-classifications show 456`,
}

func init() {
	viewCmd.AddCommand(crewRequirementCredentialClassificationsCmd)
}
