package cli

import "github.com/spf13/cobra"

var doPredictionSubjectMembershipsCmd = &cobra.Command{
	Use:   "prediction-subject-memberships",
	Short: "Manage prediction subject memberships",
	Long: `Create, update, and delete prediction subject memberships.

Prediction subject memberships link users to prediction subjects and define
what permissions they have on those subjects.

Commands:
  create    Create a prediction subject membership
  update    Update a prediction subject membership
  delete    Delete a prediction subject membership`,
	Example: `  # Create a prediction subject membership
  xbe do prediction-subject-memberships create --prediction-subject 123 --user 456

  # Update a prediction subject membership
  xbe do prediction-subject-memberships update 789 --can-manage-memberships true

  # Delete a prediction subject membership
  xbe do prediction-subject-memberships delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doPredictionSubjectMembershipsCmd)
}
