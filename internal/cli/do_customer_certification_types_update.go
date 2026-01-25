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

type doCustomerCertificationTypesUpdateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	ID                  string
	CustomerID          string
	CertificationTypeID string
}

func newDoCustomerCertificationTypesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a customer certification type",
		Long: `Update a customer certification type.

Arguments:
  <id>  The customer certification type ID (required)

Optional flags:
  --customer             Customer ID
  --certification-type   Certification type ID

Note: The certification type must belong to the customer's broker.`,
		Example: `  # Update a customer certification type
  xbe do customer-certification-types update 123 --customer 456 --certification-type 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCustomerCertificationTypesUpdate,
	}
	initDoCustomerCertificationTypesUpdateFlags(cmd)
	return cmd
}

func init() {
	doCustomerCertificationTypesCmd.AddCommand(newDoCustomerCertificationTypesUpdateCmd())
}

func initDoCustomerCertificationTypesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("customer", "", "Customer ID")
	cmd.Flags().String("certification-type", "", "Certification type ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCustomerCertificationTypesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCustomerCertificationTypesUpdateOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if strings.TrimSpace(opts.ID) == "" {
		return fmt.Errorf("customer certification type id is required")
	}

	relationships := map[string]any{}

	if cmd.Flags().Changed("customer") {
		if strings.TrimSpace(opts.CustomerID) == "" {
			return fmt.Errorf("--customer cannot be empty")
		}
		relationships["customer"] = map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.CustomerID,
			},
		}
	}

	if cmd.Flags().Changed("certification-type") {
		if strings.TrimSpace(opts.CertificationTypeID) == "" {
			return fmt.Errorf("--certification-type cannot be empty")
		}
		relationships["certification-type"] = map[string]any{
			"data": map[string]any{
				"type": "certification-types",
				"id":   opts.CertificationTypeID,
			},
		}
	}

	if len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":          "customer-certification-types",
		"id":            opts.ID,
		"relationships": relationships,
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/customer-certification-types/"+opts.ID, jsonBody)
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

	row := customerCertificationTypeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated customer certification type %s\n", row.ID)
	return nil
}

func parseDoCustomerCertificationTypesUpdateOptions(cmd *cobra.Command, args []string) (doCustomerCertificationTypesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	customerID, _ := cmd.Flags().GetString("customer")
	certificationTypeID, _ := cmd.Flags().GetString("certification-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerCertificationTypesUpdateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		ID:                  args[0],
		CustomerID:          customerID,
		CertificationTypeID: certificationTypeID,
	}, nil
}
