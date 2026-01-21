package cli

import "github.com/spf13/cobra"

var userCredentialClassificationsCmd = &cobra.Command{
	Use:   "user-credential-classifications",
	Short: "View user credential classifications",
	Long: `View user credential classifications.

These classifications define types of credentials that can be assigned to users.

Commands:
  list  List user credential classifications`,
}

func init() {
	viewCmd.AddCommand(userCredentialClassificationsCmd)
}
