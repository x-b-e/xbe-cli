package cli

import "github.com/spf13/cobra"

var laborClassificationsCmd = &cobra.Command{
	Use:   "labor-classifications",
	Short: "View labor classifications",
	Long: `View labor classifications on the XBE platform.

Labor classifications define types of workers (e.g., raker, screedman, foreman)
with their capabilities and permissions such as time card approval and project
management access.

Commands:
  list    List labor classifications`,
	Example: `  # List labor classifications
  xbe view labor-classifications list

  # Filter by name
  xbe view labor-classifications list --name "foreman"

  # Output as JSON
  xbe view labor-classifications list --json`,
}

func init() {
	viewCmd.AddCommand(laborClassificationsCmd)
}
