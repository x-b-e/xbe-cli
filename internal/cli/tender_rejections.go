package cli

import "github.com/spf13/cobra"

var tenderRejectionsCmd = &cobra.Command{
	Use:     "tender-rejections",
	Aliases: []string{"tender-rejection"},
	Short:   "View tender rejections",
	Long: `View tender rejections.

Rejections record a status change from offered to rejected for a tender and may
include a comment.

Commands:
  list    List tender rejections
  show    Show tender rejection details`,
	Example: `  # List rejections
  xbe view tender-rejections list

  # Show a rejection
  xbe view tender-rejections show 123`,
}

func init() {
	viewCmd.AddCommand(tenderRejectionsCmd)
}
