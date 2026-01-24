package cli

import "github.com/spf13/cobra"

var doTruckerApplicationApprovalsCmd = &cobra.Command{
	Use:   "trucker-application-approvals",
	Short: "Approve trucker applications",
	Long: `Approve trucker applications.

Trucker application approvals create a trucker from the application data
and mark the application as approved.

Commands:
  create    Approve a trucker application`,
	Example: `  # Approve a trucker application
  xbe do trucker-application-approvals create --trucker-application 123

  # Also add the application user as a trucker manager
  xbe do trucker-application-approvals create --trucker-application 123 --add-application-user-as-trucker-manager

  # JSON output
  xbe do trucker-application-approvals create --trucker-application 123 --json`,
}

func init() {
	doCmd.AddCommand(doTruckerApplicationApprovalsCmd)
}
