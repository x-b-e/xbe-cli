package cli

import "github.com/spf13/cobra"

var doJobSiteTimesCmd = &cobra.Command{
	Use:   "job-site-times",
	Short: "Manage job site times",
	Long: `Create, update, and delete job site times.

Job site times track how long a user spent at a job site for a job production plan.

Commands:
  create    Create a job site time
  update    Update a job site time
  delete    Delete a job site time`,
}

func init() {
	doCmd.AddCommand(doJobSiteTimesCmd)
}
