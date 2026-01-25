package cli

import "github.com/spf13/cobra"

var commentReactionsCmd = &cobra.Command{
	Use:     "comment-reactions",
	Aliases: []string{"comment-reaction"},
	Short:   "View comment reactions",
	Long: `View reactions on comments.

Comment reactions capture emoji reactions applied to comments.

Commands:
  list    List comment reactions
  show    Show comment reaction details`,
	Example: `  # List comment reactions
  xbe view comment-reactions list

  # Show a comment reaction
  xbe view comment-reactions show 123`,
}

func init() {
	viewCmd.AddCommand(commentReactionsCmd)
}
