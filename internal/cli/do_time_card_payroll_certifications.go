package cli

import "github.com/spf13/cobra"

var doTimeCardPayrollCertificationsCmd = &cobra.Command{
	Use:   "time-card-payroll-certifications",
	Short: "Manage time card payroll certifications",
	Long: `Create and delete time card payroll certifications.

Time card payroll certifications indicate a time card has been certified for payroll.

Commands:
  create  Create a time card payroll certification
  delete  Delete a time card payroll certification`,
	Example: `  # Certify a time card for payroll
  xbe do time-card-payroll-certifications create --time-card 123

  # Delete a certification
  xbe do time-card-payroll-certifications delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doTimeCardPayrollCertificationsCmd)
}
