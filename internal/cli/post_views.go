package cli

import "github.com/spf13/cobra"

var postViewsCmd = &cobra.Command{
	Use:     "post-views",
	Aliases: []string{"post-view"},
	Short:   "Browse post views",
	Long: `Browse post views.

Post views record when a user views a post.

Commands:
  list    List post views
  show    Show post view details`,
	Example: `  # List post views
  xbe view post-views list

  # Show a post view
  xbe view post-views show 123

  # Output as JSON
  xbe view post-views list --json`,
}

func init() {
	viewCmd.AddCommand(postViewsCmd)
}
