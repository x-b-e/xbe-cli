package cli

import "github.com/spf13/cobra"

var jobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "View jobs",
	Long: `View jobs on the XBE platform.

Jobs tie together customers, job sites, material types, and trailer classifications.
They can optionally be linked to job production plans.

Commands:
  list    List jobs with filtering
  show    Show job details`,
	Example: `  # List jobs
  xbe view jobs list

  # Filter by customer
  xbe view jobs list --customer 123

  # Show a job
  xbe view jobs show 456`,
}

func init() {
	viewCmd.AddCommand(jobsCmd)
}
