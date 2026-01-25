package cli

import "github.com/spf13/cobra"

var postRouterJobsCmd = &cobra.Command{
	Use:     "post-router-jobs",
	Aliases: []string{"post-router-job"},
	Short:   "Browse post router jobs",
	Long: `Browse post router jobs.

Post router jobs track background worker jobs created for routed posts.

Commands:
  list    List post router jobs with filtering and pagination
  show    Show full details of a post router job`,
	Example: `  # List post router jobs
  xbe view post-router-jobs list

  # Filter by post router
  xbe view post-router-jobs list --post-router 123

  # Filter by post
  xbe view post-router-jobs list --post 456

  # Filter by worker class
  xbe view post-router-jobs list --post-worker-class-name "Posters::FooWorker"

  # Show post router job details
  xbe view post-router-jobs show 789`,
}

func init() {
	viewCmd.AddCommand(postRouterJobsCmd)
}
