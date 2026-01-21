package cli

import "github.com/spf13/cobra"

var stakeholderClassificationsCmd = &cobra.Command{
	Use:   "stakeholder-classifications",
	Short: "View stakeholder classifications",
	Long: `View stakeholder classifications on the XBE platform.

Stakeholder classifications categorize project stakeholders by their role
and influence level (leverage factor).

Commands:
  list    List stakeholder classifications`,
	Example: `  # List stakeholder classifications
  xbe view stakeholder-classifications list

  # Filter by slug
  xbe view stakeholder-classifications list --slug "owner"

  # Filter by leverage factor
  xbe view stakeholder-classifications list --leverage-factor 5

  # Output as JSON
  xbe view stakeholder-classifications list --json`,
}

func init() {
	viewCmd.AddCommand(stakeholderClassificationsCmd)
}
