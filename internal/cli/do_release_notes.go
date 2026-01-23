package cli

import "github.com/spf13/cobra"

var doReleaseNotesCmd = &cobra.Command{
	Use:     "release-notes",
	Aliases: []string{"release-note"},
	Short:   "Manage release notes",
	Long:    `Create release notes.`,
}

func init() {
	doCmd.AddCommand(doReleaseNotesCmd)
}
