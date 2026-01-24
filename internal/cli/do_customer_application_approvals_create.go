package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doCustomerApplicationApprovalsCreateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	CustomerApplicationID string
	CreditLimit           string
}

type customerApplicationApprovalRow struct {
	ID                    string `json:"id"`
	CustomerApplicationID string `json:"customer_application_id,omitempty"`
	CustomerID            string `json:"customer_id,omitempty"`
	CreditLimit           string `json:"credit_limit,omitempty"`
}

func newDoCustomerApplicationApprovalsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Approve a customer application",
		Long: `Approve a customer application.

Approvals create a customer from the application data and set its credit limit.

Required flags:
  --customer-application   Customer application ID
  --credit-limit           Credit limit for the new customer`,
		Example: `  # Approve a customer application
  xbe do customer-application-approvals create --customer-application 123 --credit-limit 1000000

  # JSON output
  xbe do customer-application-approvals create --customer-application 123 --credit-limit 1000000 --json`,
		RunE: runDoCustomerApplicationApprovalsCreate,
	}
	initDoCustomerApplicationApprovalsCreateFlags(cmd)
	return cmd
}

func init() {
	doCustomerApplicationApprovalsCmd.AddCommand(newDoCustomerApplicationApprovalsCreateCmd())
}

func initDoCustomerApplicationApprovalsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("customer-application", "", "Customer application ID (required)")
	cmd.Flags().String("credit-limit", "", "Credit limit for the customer (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("customer-application")
	cmd.MarkFlagRequired("credit-limit")
}

func runDoCustomerApplicationApprovalsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCustomerApplicationApprovalsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	attributes := map[string]any{
		"customer-application-id": opts.CustomerApplicationID,
		"credit-limit":            opts.CreditLimit,
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "customer-application-approvals",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/customer-application-approvals", jsonBody)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	row := buildCustomerApplicationApprovalRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created customer application approval %s\n", row.ID)
	return nil
}

func parseDoCustomerApplicationApprovalsCreateOptions(cmd *cobra.Command) (doCustomerApplicationApprovalsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	customerApplicationID, _ := cmd.Flags().GetString("customer-application")
	creditLimit, _ := cmd.Flags().GetString("credit-limit")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerApplicationApprovalsCreateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		CustomerApplicationID: customerApplicationID,
		CreditLimit:           creditLimit,
	}, nil
}

func buildCustomerApplicationApprovalRowFromSingle(resp jsonAPISingleResponse) customerApplicationApprovalRow {
	resource := resp.Data
	row := customerApplicationApprovalRow{
		ID:                    resource.ID,
		CustomerApplicationID: stringAttr(resource.Attributes, "customer-application-id"),
		CustomerID:            stringAttr(resource.Attributes, "customer-id"),
		CreditLimit:           stringAttr(resource.Attributes, "credit-limit"),
	}
	if row.CustomerApplicationID == "" {
		row.CustomerApplicationID = resource.ID
	}
	return row
}
