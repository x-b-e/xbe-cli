package cli

import "github.com/spf13/cobra"

var objectiveChangesCmd = &cobra.Command{
	Use:     "objective-changes",
	Aliases: []string{"objective-change"},
	Short:   "View objective changes",
	Long: `View objective changes on the XBE platform.

Objective changes capture updates to objective schedule dates.

Commands:
  list    List objective changes
  show    Show objective change details`,
	Example: `  # List objective changes
  xbe view objective-changes list

  # Filter by objective
  xbe view objective-changes list --objective 123

  # Show an objective change
  xbe view objective-changes show 456`,
}

func init() {
	viewCmd.AddCommand(objectiveChangesCmd)
}
