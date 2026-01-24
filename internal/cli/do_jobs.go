package cli

import "github.com/spf13/cobra"

var doJobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Manage jobs",
	Long: `Create, update, and delete jobs.

Jobs tie together customers, job sites, material types, and trailer classifications.
They can optionally be linked to job production plans.

Commands:
  create    Create a new job
  update    Update an existing job
  delete    Delete a job`,
}

func init() {
	doCmd.AddCommand(doJobsCmd)
}
