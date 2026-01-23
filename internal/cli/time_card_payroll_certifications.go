package cli

import "github.com/spf13/cobra"

var timeCardPayrollCertificationsCmd = &cobra.Command{
	Use:   "time-card-payroll-certifications",
	Short: "View time card payroll certifications",
	Long: `View time card payroll certifications.

Time card payroll certifications indicate a time card has been certified for payroll.

Commands:
  list    List time card payroll certifications
  show    Show time card payroll certification details`,
	Example: `  # List time card payroll certifications
  xbe view time-card-payroll-certifications list

  # Show a time card payroll certification
  xbe view time-card-payroll-certifications show 123`,
}

func init() {
	viewCmd.AddCommand(timeCardPayrollCertificationsCmd)
}
