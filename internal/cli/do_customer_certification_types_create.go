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

type doCustomerCertificationTypesCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	CustomerID          string
	CertificationTypeID string
}

func newDoCustomerCertificationTypesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a customer certification type",
		Long: `Create a customer certification type.

Required flags:
  --customer             Customer ID (required)
  --certification-type   Certification type ID (required)

Note: The certification type must belong to the customer's broker.`,
		Example: `  # Create a customer certification type
  xbe do customer-certification-types create --customer 123 --certification-type 456`,
		Args: cobra.NoArgs,
		RunE: runDoCustomerCertificationTypesCreate,
	}
	initDoCustomerCertificationTypesCreateFlags(cmd)
	return cmd
}

func init() {
	doCustomerCertificationTypesCmd.AddCommand(newDoCustomerCertificationTypesCreateCmd())
}

func initDoCustomerCertificationTypesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("customer", "", "Customer ID (required)")
	cmd.Flags().String("certification-type", "", "Certification type ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("customer")
	cmd.MarkFlagRequired("certification-type")
}

func runDoCustomerCertificationTypesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCustomerCertificationTypesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.CustomerID) == "" {
		err := fmt.Errorf("--customer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.CertificationTypeID) == "" {
		err := fmt.Errorf("--certification-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"customer": map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.CustomerID,
			},
		},
		"certification-type": map[string]any{
			"data": map[string]any{
				"type": "certification-types",
				"id":   opts.CertificationTypeID,
			},
		},
	}

	data := map[string]any{
		"type":          "customer-certification-types",
		"relationships": relationships,
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/customer-certification-types", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created customer certification type %s\n", row.ID)
	return nil
}

func parseDoCustomerCertificationTypesCreateOptions(cmd *cobra.Command) (doCustomerCertificationTypesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	customerID, _ := cmd.Flags().GetString("customer")
	certificationTypeID, _ := cmd.Flags().GetString("certification-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerCertificationTypesCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		CustomerID:          customerID,
		CertificationTypeID: certificationTypeID,
	}, nil
}
