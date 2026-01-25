package cli

import "github.com/spf13/cobra"

var mechanicUserAssociationsCmd = &cobra.Command{
	Use:     "mechanic-user-associations",
	Aliases: []string{"mechanic-user-association"},
	Short:   "Browse mechanic user associations",
	Long: `Browse mechanic user associations.

Mechanic user associations link users to maintenance requirements.

Commands:
  list    List records with filters
  show    Show record details`,
	Example: `  # List records
  xbe view mechanic-user-associations list

  # Show record details
  xbe view mechanic-user-associations show 123`,
}

func init() {
	viewCmd.AddCommand(mechanicUserAssociationsCmd)
}
