package cli

import "github.com/spf13/cobra"

var postActionsCmd = &cobra.Command{
	Use:     "post-actions",
	Aliases: []string{"post-action"},
	Short:   "Browse post actions",
	Long: `Browse post actions.

Post actions record tokenized actions used in post workflows.

Commands:
  list    List post actions
  show    Show post action details`,
	Example: `  # List post actions
  xbe view post-actions list

  # Show a post action
  xbe view post-actions show 123

  # Output as JSON
  xbe view post-actions list --json`,
}

func init() {
	viewCmd.AddCommand(postActionsCmd)
}
