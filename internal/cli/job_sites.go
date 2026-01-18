package cli

import "github.com/spf13/cobra"

var jobSitesCmd = &cobra.Command{
	Use:   "job-sites",
	Short: "View job sites",
	Long: `View job sites on the XBE platform.

Job sites are delivery locations for materials - the destinations where
trucks deliver to on job production plans.

Commands:
  list    List job sites with filtering`,
	Example: `  # List job sites
  xbe view job-sites list

  # Search by name
  xbe view job-sites list --name "Main Street"

  # List active job sites only
  xbe view job-sites list --active`,
}

func init() {
	viewCmd.AddCommand(jobSitesCmd)
}
