package cli

import "github.com/spf13/cobra"

var rootCausesCmd = &cobra.Command{
	Use:     "root-causes",
	Aliases: []string{"root-cause"},
	Short:   "View root causes",
	Long: `View root causes tied to incidents.

Root causes track underlying issues for incidents and can be linked
hierarchically to group related causes.

Commands:
  list    List root causes
  show    Show root cause details`,
	Example: `  # List root causes
  xbe view root-causes list

  # Filter by incident
  xbe view root-causes list --incident-type production-incidents --incident-id 123

  # Show a root cause
  xbe view root-causes show 456

  # Output JSON
  xbe view root-causes list --json`,
}

func init() {
	viewCmd.AddCommand(rootCausesCmd)
}
