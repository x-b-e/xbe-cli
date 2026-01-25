package cli

import "github.com/spf13/cobra"

var truckerApplicationsCmd = &cobra.Command{
	Use:     "trucker-applications",
	Aliases: []string{"trucker-application"},
	Short:   "Browse trucker applications",
	Long: `Browse trucker applications on the XBE platform.

Trucker applications represent prospective trucking companies applying to work
with a broker.

Commands:
  list    List trucker applications with filtering
  show    Show trucker application details`,
	Example: `  # List trucker applications
  xbe view trucker-applications list

  # Show a trucker application
  xbe view trucker-applications show 123`,
}

func init() {
	viewCmd.AddCommand(truckerApplicationsCmd)
}
