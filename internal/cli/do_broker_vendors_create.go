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

type doBrokerVendorsCreateOptions struct {
	BaseURL                          string
	Token                            string
	JSON                             bool
	Broker                           string
	Vendor                           string
	ExternalAccountingBrokerVendorID string
}

func newDoBrokerVendorsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a broker-vendor relationship",
		Long: `Create a broker-vendor trading partner relationship.

Required:
  --broker   Broker ID
  --vendor   Vendor in Type|ID format (Trucker or MaterialSite)

Optional:
  --external-accounting-broker-vendor-id  External accounting broker vendor ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a broker-vendor relationship for a trucker
  xbe do broker-vendors create --broker 123 --vendor "Trucker|456"

  # Create with external accounting ID
  xbe do broker-vendors create --broker 123 --vendor "MaterialSite|789" \
    --external-accounting-broker-vendor-id "ACCT-42"`,
		Args: cobra.NoArgs,
		RunE: runDoBrokerVendorsCreate,
	}
	initDoBrokerVendorsCreateFlags(cmd)
	return cmd
}

func init() {
	doBrokerVendorsCmd.AddCommand(newDoBrokerVendorsCreateCmd())
}

func initDoBrokerVendorsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("vendor", "", "Vendor in Type|ID format (Trucker or MaterialSite)")
	cmd.Flags().String("external-accounting-broker-vendor-id", "", "External accounting broker vendor ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("broker")
	_ = cmd.MarkFlagRequired("vendor")
}

func runDoBrokerVendorsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoBrokerVendorsCreateOptions(cmd)
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

	if opts.Broker == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Vendor == "" {
		err := fmt.Errorf("--vendor is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	vendorType, vendorID, err := parseBrokerVendorPartner(opts.Vendor)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.ExternalAccountingBrokerVendorID != "" {
		attributes["external-accounting-broker-vendor-id"] = opts.ExternalAccountingBrokerVendorID
	}

	relationships := map[string]any{
		"organization": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
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
			"type":          "broker-vendors",
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

	body, _, err := client.Post(cmd.Context(), "/v1/broker-vendors", jsonBody)
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

	row := buildBrokerVendorRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created broker vendor %s\n", row.ID)
	return nil
}

func parseDoBrokerVendorsCreateOptions(cmd *cobra.Command) (doBrokerVendorsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	broker, _ := cmd.Flags().GetString("broker")
	vendor, _ := cmd.Flags().GetString("vendor")
	externalAccountingBrokerVendorID, _ := cmd.Flags().GetString("external-accounting-broker-vendor-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerVendorsCreateOptions{
		BaseURL:                          baseURL,
		Token:                            token,
		JSON:                             jsonOut,
		Broker:                           broker,
		Vendor:                           vendor,
		ExternalAccountingBrokerVendorID: externalAccountingBrokerVendorID,
	}, nil
}

func parseBrokerVendorPartner(vendor string) (string, string, error) {
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
	case "materialsite", "material-site", "material_site", "material-sites":
		jsonAPIType = "material-sites"
	default:
		return "", "", fmt.Errorf("unsupported vendor type %q (expected Trucker or MaterialSite)", vendorType)
	}

	return jsonAPIType, vendorID, nil
}

func buildBrokerVendorRowFromSingle(resp jsonAPISingleResponse) brokerVendorRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	row := brokerVendorRow{
		ID:                               resp.Data.ID,
		ExternalAccountingBrokerVendorID: stringAttr(resp.Data.Attributes, "external-accounting-broker-vendor-id"),
		TradingPartnerType:               stringAttr(resp.Data.Attributes, "trading-partner-type"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		row.BrokerName = brokerVendorNameFromIncluded(rel.Data, included)
	} else if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		row.BrokerName = brokerVendorNameFromIncluded(rel.Data, included)
	}

	if rel, ok := resp.Data.Relationships["vendor"]; ok && rel.Data != nil {
		row.VendorID = rel.Data.ID
		row.VendorName = brokerVendorNameFromIncluded(rel.Data, included)
	} else if rel, ok := resp.Data.Relationships["partner"]; ok && rel.Data != nil {
		row.VendorID = rel.Data.ID
		row.VendorName = brokerVendorNameFromIncluded(rel.Data, included)
	}

	return row
}
