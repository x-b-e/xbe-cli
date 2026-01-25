package cli

import "github.com/spf13/cobra"

var predictionSubjectMembershipsCmd = &cobra.Command{
	Use:   "prediction-subject-memberships",
	Short: "Browse prediction subject memberships",
	Long: `Browse prediction subject memberships.

Prediction subject memberships link users to prediction subjects and define
what permissions they have on those subjects.

Commands:
  list    List prediction subject memberships with filtering
  show    Show prediction subject membership details`,
	Example: `  # List prediction subject memberships
  xbe view prediction-subject-memberships list

  # Filter by prediction subject
  xbe view prediction-subject-memberships list --prediction-subject 123

  # Show a prediction subject membership
  xbe view prediction-subject-memberships show 456`,
}

func init() {
	viewCmd.AddCommand(predictionSubjectMembershipsCmd)
}
