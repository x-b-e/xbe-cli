package cli

import "github.com/spf13/cobra"

var doShiftScopeMatchesCmd = &cobra.Command{
	Use:     "shift-scope-matches",
	Aliases: []string{"shift-scope-match"},
	Short:   "Match shift scopes against tenders",
	Long: `Match shift scopes against tenders.

Commands:
  create    Match shift scopes`,
	Example: `  # Match a tender against a rate
  xbe do shift-scope-matches create --tender 123 --rate 456

  # Match using a shift set time card constraint
  xbe do shift-scope-matches create --tender 123 --shift-set-time-card-constraint 789

  # JSON output
  xbe do shift-scope-matches create --tender 123 --rate 456 --json`,
}

func init() {
	doCmd.AddCommand(doShiftScopeMatchesCmd)
}
