package cli

import "github.com/spf13/cobra"

var projectDivisionsCmd = &cobra.Command{
	Use:   "project-divisions",
	Short: "View project divisions",
	Long: `View project divisions on the XBE platform.

Project divisions are organizational units used to group projects by business
division or department. They help segment reporting and control access.

Commands:
  list    List project divisions`,
	Example: `  # List project divisions
  xbe view project-divisions list

  # Search by name
  xbe view project-divisions list --name "North"`,
}

func init() {
	viewCmd.AddCommand(projectDivisionsCmd)
}
