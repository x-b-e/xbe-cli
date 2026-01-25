package cli

import "github.com/spf13/cobra"

var timeCardApprovalAuditsCmd = &cobra.Command{
	Use:   "time-card-approval-audits",
	Short: "View time card approval audits",
	Long: `View time card approval audits.

Time card approval audits track who audited a time card and when.

Commands:
  list    List time card approval audits
  show    Show time card approval audit details`,
	Example: `  # List time card approval audits
  xbe view time-card-approval-audits list

  # Show a time card approval audit
  xbe view time-card-approval-audits show 123`,
}

func init() {
	viewCmd.AddCommand(timeCardApprovalAuditsCmd)
}
