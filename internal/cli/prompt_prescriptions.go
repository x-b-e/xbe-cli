package cli

import "github.com/spf13/cobra"

var promptPrescriptionsCmd = &cobra.Command{
	Use:     "prompt-prescriptions",
	Aliases: []string{"prompt-prescription"},
	Short:   "Browse prompt prescriptions",
	Long: `Browse prompt prescriptions.

Prompt prescriptions capture user details and generate tailored AI prompt
suggestions for heavy materials and construction workflows.

Commands:
  list    List prompt prescriptions
  show    Show prompt prescription details`,
	Example: `  # List prompt prescriptions
  xbe view prompt-prescriptions list

  # Filter by email address
  xbe view prompt-prescriptions list --email-address "name@example.com"

  # Show a prompt prescription
  xbe view prompt-prescriptions show 123`,
}

func init() {
	viewCmd.AddCommand(promptPrescriptionsCmd)
}
