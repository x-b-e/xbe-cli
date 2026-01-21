package cli

import "github.com/spf13/cobra"

var doStakeholderClassificationsCmd = &cobra.Command{
	Use:   "stakeholder-classifications",
	Short: "Manage stakeholder classifications",
	Long: `Create, update, and delete stakeholder classifications.

Stakeholder classifications categorize project stakeholders by their role
and influence level (leverage factor).

Note: Only admin users can create, update, or delete stakeholder classifications.

Commands:
  create    Create a new stakeholder classification
  update    Update an existing stakeholder classification
  delete    Delete a stakeholder classification`,
	Example: `  # Create a stakeholder classification
  xbe do stakeholder-classifications create --title "Project Owner" --leverage-factor 5

  # Update a stakeholder classification
  xbe do stakeholder-classifications update 123 --title "Primary Owner"

  # Delete a stakeholder classification
  xbe do stakeholder-classifications delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doStakeholderClassificationsCmd)
}
