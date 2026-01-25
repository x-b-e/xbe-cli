package cli

import "github.com/spf13/cobra"

var doRateAgreementCopiersCmd = &cobra.Command{
	Use:   "rate-agreement-copiers",
	Short: "Copy rate agreements to target organizations",
	Long: `Copy rate agreements to target organizations.

Rate agreement copiers create or update a target rate agreement using a template
rate agreement. The target organization must be a customer or trucker that
matches the template buyer/seller role.

Commands:
  create    Copy a rate agreement to a target organization`,
	Example: `  # Copy a template rate agreement to a customer
  xbe do rate-agreement-copiers create \
    --template-rate-agreement 123 \
    --target-organization-type customers \
    --target-organization-id 456

  # Copy a template rate agreement to a trucker
  xbe do rate-agreement-copiers create \
    --template-rate-agreement 123 \
    --target-organization-type truckers \
    --target-organization-id 789`,
}

func init() {
	doCmd.AddCommand(doRateAgreementCopiersCmd)
}
