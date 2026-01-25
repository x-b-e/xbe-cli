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

type doBusinessUnitCustomersCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	BusinessUnit string
	Customer     string
}

func newDoBusinessUnitCustomersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a business unit customer link",
		Long: `Create a business unit customer link.

Required flags:
  --business-unit  Business unit ID (required)
  --customer       Customer ID (required)`,
		Example: `  # Link a business unit to a customer
  xbe do business-unit-customers create \\
    --business-unit 123 \\
    --customer 456

  # JSON output
  xbe do business-unit-customers create \\
    --business-unit 123 \\
    --customer 456 \\
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoBusinessUnitCustomersCreate,
	}
	initDoBusinessUnitCustomersCreateFlags(cmd)
	return cmd
}

func init() {
	doBusinessUnitCustomersCmd.AddCommand(newDoBusinessUnitCustomersCreateCmd())
}

func initDoBusinessUnitCustomersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("business-unit", "", "Business unit ID (required)")
	cmd.Flags().String("customer", "", "Customer ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBusinessUnitCustomersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoBusinessUnitCustomersCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if strings.TrimSpace(opts.BusinessUnit) == "" {
		err := fmt.Errorf("--business-unit is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Customer) == "" {
		err := fmt.Errorf("--customer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"business-unit": map[string]any{
			"data": map[string]any{
				"type": "business-units",
				"id":   opts.BusinessUnit,
			},
		},
		"customer": map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.Customer,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "business-unit-customers",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/business-unit-customers", jsonBody)
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

	row := buildBusinessUnitCustomerRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created business unit customer %s\n", row.ID)
	return nil
}

func parseDoBusinessUnitCustomersCreateOptions(cmd *cobra.Command) (doBusinessUnitCustomersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	customer, _ := cmd.Flags().GetString("customer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBusinessUnitCustomersCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		BusinessUnit: businessUnit,
		Customer:     customer,
	}, nil
}

func buildBusinessUnitCustomerRowFromSingle(resp jsonAPISingleResponse) businessUnitCustomerRow {
	row := businessUnitCustomerRow{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["business-unit"]; ok && rel.Data != nil {
		row.BusinessUnitID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["customer"]; ok && rel.Data != nil {
		row.CustomerID = rel.Data.ID
	}

	return row
}
