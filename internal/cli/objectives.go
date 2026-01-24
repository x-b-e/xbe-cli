package cli

import "github.com/spf13/cobra"

var objectivesCmd = &cobra.Command{
	Use:     "objectives",
	Aliases: []string{"objective"},
	Short:   "View objectives",
	Long: `View objectives on the XBE platform.

Objectives capture goals, timelines, and ownership for key initiatives.

Commands:
  list    List objectives
  show    Show objective details`,
	Example: `  # List objectives
  xbe view objectives list

  # Filter by owner
  xbe view objectives list --owner 123

  # Show an objective
  xbe view objectives show 456`,
}

func init() {
	viewCmd.AddCommand(objectivesCmd)
}
