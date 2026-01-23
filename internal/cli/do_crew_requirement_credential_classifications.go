package cli

import "github.com/spf13/cobra"

var doCrewRequirementCredentialClassificationsCmd = &cobra.Command{
	Use:     "crew-requirement-credential-classifications",
	Aliases: []string{"crew-requirement-credential-classification"},
	Short:   "Manage crew requirement credential classifications",
	Long:    "Commands for creating and deleting crew requirement credential classifications.",
}

func init() {
	doCmd.AddCommand(doCrewRequirementCredentialClassificationsCmd)
}
