package cli

import "github.com/spf13/cobra"

var maintenanceCmd = &cobra.Command{
	Use:   "maintenance",
	Short: "Browse maintenance requirements, sets, and rules",
	Long: `Browse maintenance requirements, sets, and rules on the XBE platform.

Maintenance commands provide access to equipment maintenance scheduling,
requirement tracking, and rule configuration.

Commands:
  requirements  View maintenance requirements
  sets          View maintenance requirement sets
  rules         View maintenance requirement rules
  parts         View maintenance requirement parts catalog`,
	Example: `  # List maintenance requirements
  xbe view maintenance requirements list --status pending

  # List requirements for my business units
  xbe view maintenance requirements list --me

  # List requirement sets
  xbe view maintenance sets list --status in_progress

  # List sets for my business units
  xbe view maintenance sets list --me

  # List active rules
  xbe view maintenance rules list --active-only

  # List rules for my business units
  xbe view maintenance rules list --me

  # List parts catalog
  xbe view maintenance parts list`,
}

func init() {
	viewCmd.AddCommand(maintenanceCmd)
}
