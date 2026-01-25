package cli

import "github.com/spf13/cobra"

var doHosAnnotationsCmd = &cobra.Command{
	Use:     "hos-annotations",
	Aliases: []string{"hos-annotation"},
	Short:   "Manage HOS annotations",
	Long: `Commands for managing HOS annotations.

HOS annotations are read-only in most contexts; deletion requires admin access.`,
}

func init() {
	doCmd.AddCommand(doHosAnnotationsCmd)
}
