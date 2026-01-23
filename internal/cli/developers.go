package cli

import "github.com/spf13/cobra"

var developersCmd = &cobra.Command{
	Use:   "developers",
	Short: "View developers",
	Long: `View developers on the XBE platform.

Developers are companies that develop projects. They are distinct from
customers and have their own set of projects, certifications, and reference types.

Commands:
  list    List developers`,
	Example: `  # List developers
  xbe view developers list

  # Search by name
  xbe view developers list --name "Acme"

  # Filter by broker
  xbe view developers list --broker 123`,
}

func init() {
	viewCmd.AddCommand(developersCmd)
}
