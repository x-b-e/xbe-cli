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

type doCustomerVendorsCreateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	Customer                           string
	Vendor                             string
	ExternalAccountingCustomerVendorID string
}

func newDoCustomerVendorsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a customer-vendor relationship",
		Long: `Create a customer-vendor trading partner relationship.

Required:
  --customer  Customer ID
  --vendor    Vendor in Type|ID format (Trucker)

Optional:
  --external-accounting-customer-vendor-id  External accounting customer vendor ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a customer-vendor relationship for a trucker
  xbe do customer-vendors create --customer 123 --vendor "Trucker|456"

  # Create with external accounting ID
  xbe do customer-vendors create --customer 123 --vendor "Trucker|456" \
    --external-accounting-customer-vendor-id "ACCT-42"`,
		Args: cobra.NoArgs,
		RunE: runDoCustomerVendorsCreate,
	}
	initDoCustomerVendorsCreateFlags(cmd)
	return cmd
}

func init() {
	doCustomerVendorsCmd.AddCommand(newDoCustomerVendorsCreateCmd())
}

func initDoCustomerVendorsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("customer", "", "Customer ID")
	cmd.Flags().String("vendor", "", "Vendor in Type|ID format (Trucker)")
	cmd.Flags().String("external-accounting-customer-vendor-id", "", "External accounting customer vendor ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("customer")
	_ = cmd.MarkFlagRequired("vendor")
}

func runDoCustomerVendorsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCustomerVendorsCreateOptions(cmd)
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

	if opts.Customer == "" {
		err := fmt.Errorf("--customer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Vendor == "" {
		err := fmt.Errorf("--vendor is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	vendorType, vendorID, err := parseCustomerVendorPartner(opts.Vendor)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.ExternalAccountingCustomerVendorID != "" {
		attributes["external-accounting-customer-vendor-id"] = opts.ExternalAccountingCustomerVendorID
	}

	relationships := map[string]any{
		"organization": map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.Customer,
			},
		},
		"partner": map[string]any{
			"data": map[string]any{
				"type": vendorType,
				"id":   vendorID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "customer-vendors",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/customer-vendors", jsonBody)
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

	row := buildCustomerVendorRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created customer vendor %s\n", row.ID)
	return nil
}

func parseDoCustomerVendorsCreateOptions(cmd *cobra.Command) (doCustomerVendorsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	customer, _ := cmd.Flags().GetString("customer")
	vendor, _ := cmd.Flags().GetString("vendor")
	externalAccountingCustomerVendorID, _ := cmd.Flags().GetString("external-accounting-customer-vendor-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerVendorsCreateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		Customer:                           customer,
		Vendor:                             vendor,
		ExternalAccountingCustomerVendorID: externalAccountingCustomerVendorID,
	}, nil
}

func parseCustomerVendorPartner(vendor string) (string, string, error) {
	parts := strings.SplitN(vendor, "|", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid vendor format: %q (expected Type|ID, e.g. Trucker|123)", vendor)
	}
	vendorType := strings.TrimSpace(parts[0])
	vendorID := strings.TrimSpace(parts[1])
	if vendorType == "" || vendorID == "" {
		return "", "", fmt.Errorf("invalid vendor format: %q (expected Type|ID, e.g. Trucker|123)", vendor)
	}

	jsonAPIType := strings.ToLower(vendorType)
	switch jsonAPIType {
	case "trucker", "truckers":
		jsonAPIType = "truckers"
	default:
		return "", "", fmt.Errorf("unsupported vendor type %q (expected Trucker)", vendorType)
	}

	return jsonAPIType, vendorID, nil
}

func buildCustomerVendorRowFromSingle(resp jsonAPISingleResponse) customerVendorRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	row := customerVendorRow{
		ID:                                 resp.Data.ID,
		ExternalAccountingCustomerVendorID: stringAttr(resp.Data.Attributes, "external-accounting-customer-vendor-id"),
		TradingPartnerType:                 stringAttr(resp.Data.Attributes, "trading-partner-type"),
	}

	if rel, ok := resp.Data.Relationships["customer"]; ok && rel.Data != nil {
		row.CustomerID = rel.Data.ID
		row.CustomerName = customerVendorNameFromIncluded(rel.Data, included)
	} else if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		row.CustomerID = rel.Data.ID
		row.CustomerName = customerVendorNameFromIncluded(rel.Data, included)
	}

	if rel, ok := resp.Data.Relationships["vendor"]; ok && rel.Data != nil {
		row.VendorID = rel.Data.ID
		row.VendorName = customerVendorNameFromIncluded(rel.Data, included)
	} else if rel, ok := resp.Data.Relationships["partner"]; ok && rel.Data != nil {
		row.VendorID = rel.Data.ID
		row.VendorName = customerVendorNameFromIncluded(rel.Data, included)
	}

	return row
}
