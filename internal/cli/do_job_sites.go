package cli

import "github.com/spf13/cobra"

var doJobSitesCmd = &cobra.Command{
	Use:   "job-sites",
	Short: "Manage job sites",
	Long: `Create, update, and delete job sites.

Job sites are delivery locations for materials - the destinations where
trucks deliver to on job production plans.

Commands:
  create    Create a new job site
  update    Update an existing job site
  delete    Delete a job site`,
}

func init() {
	doCmd.AddCommand(doJobSitesCmd)
}
