package cli

import "github.com/spf13/cobra"

var postRoutersCmd = &cobra.Command{
	Use:     "post-routers",
	Aliases: []string{"post-router"},
	Short:   "Browse post routers",
	Long: `Browse post routers.

Post routers analyze posts and queue routing jobs.

Commands:
  list    List post routers with filtering and pagination
  show    Show full details of a post router`,
	Example: `  # List post routers
  xbe view post-routers list

  # Filter by status
  xbe view post-routers list --status queueing

  # Filter by post
  xbe view post-routers list --post 123

  # Show post router details
  xbe view post-routers show 456`,
}

func init() {
	viewCmd.AddCommand(postRoutersCmd)
}
