package cli

import "github.com/spf13/cobra"

var doTimeCardApprovalAuditsCmd = &cobra.Command{
	Use:   "time-card-approval-audits",
	Short: "Manage time card approval audits",
	Long: `Create, update, and delete time card approval audits.

Time card approval audits record who audited an approved time card. Audits are
unique per time card and can be deleted only when the time card has no invoices.
Only admins or time card auditors can manage audits.

Commands:
  create  Create a new approval audit
  update  Update an existing approval audit
  delete  Delete an approval audit`,
	Example: `  # Create a time card approval audit
  xbe do time-card-approval-audits create --time-card 123 --user 456 --note "Reviewed"

  # Update an audit note
  xbe do time-card-approval-audits update 789 --note "Updated note"

  # Delete an audit
  xbe do time-card-approval-audits delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doTimeCardApprovalAuditsCmd)
}
