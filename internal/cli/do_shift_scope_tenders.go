package cli

import "github.com/spf13/cobra"

var doShiftScopeTendersCmd = &cobra.Command{
	Use:     "shift-scope-tenders",
	Aliases: []string{"shift-scope-tender"},
	Short:   "Find tenders for a shift scope",
	Long: `Find tenders for a shift scope.

Commands:
  create    Find tenders for a shift scope`,
}

func init() {
	doCmd.AddCommand(doShiftScopeTendersCmd)
}
