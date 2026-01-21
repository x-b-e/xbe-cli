package cli

import "github.com/spf13/cobra"

var projectOfficesCmd = &cobra.Command{
	Use:   "project-offices",
	Short: "View project offices",
	Long: `View project offices on the XBE platform.

Project offices are geographic or organizational divisions used to group
projects and transport orders. They help organize work by region or office
location.

Commands:
  list    List project offices`,
	Example: `  # List project offices
  xbe view project-offices list

  # Search by name
  xbe view project-offices list --name "Chicago"

  # Filter by broker
  xbe view project-offices list --broker 123`,
}

func init() {
	viewCmd.AddCommand(projectOfficesCmd)
}
