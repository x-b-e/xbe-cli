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

type doTradingPartnersCreateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	Organization       string
	Partner            string
	TradingPartnerType string
}

func newDoTradingPartnersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a trading partner link",
		Long: `Create a trading partner link between an organization and a partner.

Required:
  --organization  Organization in Type|ID format (Broker or Customer)
  --partner       Partner in Type|ID format (Customer, Broker, Trucker, MaterialSite)

Optional:
  --trading-partner-type  Trading partner type (e.g. BrokerCustomer, BrokerVendor, CustomerVendor)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a broker/customer trading partner
  xbe do trading-partners create --organization "Broker|123" --partner "Customer|456" \
    --trading-partner-type BrokerCustomer

  # Create a broker/trucker trading partner
  xbe do trading-partners create --organization "Broker|123" --partner "Trucker|456"

  # JSON output
  xbe do trading-partners create --organization "Customer|123" --partner "Trucker|456" --json`,
		Args: cobra.NoArgs,
		RunE: runDoTradingPartnersCreate,
	}
	initDoTradingPartnersCreateFlags(cmd)
	return cmd
}

func init() {
	doTradingPartnersCmd.AddCommand(newDoTradingPartnersCreateCmd())
}

func initDoTradingPartnersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("organization", "", "Organization in Type|ID format (Broker or Customer)")
	cmd.Flags().String("partner", "", "Partner in Type|ID format (Customer, Broker, Trucker, MaterialSite)")
	cmd.Flags().String("trading-partner-type", "", "Trading partner type (e.g. BrokerCustomer)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("organization")
	_ = cmd.MarkFlagRequired("partner")
}

func runDoTradingPartnersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTradingPartnersCreateOptions(cmd)
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

	if opts.Organization == "" {
		err := fmt.Errorf("--organization is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Partner == "" {
		err := fmt.Errorf("--partner is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	organizationType, organizationID, err := parseTradingPartnerReference(opts.Organization, "organization", map[string]string{
		"brokers":   "Broker",
		"customers": "Customer",
	})
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	partnerType, partnerID, err := parseTradingPartnerReference(opts.Partner, "partner", map[string]string{
		"customers":      "Customer",
		"brokers":        "Broker",
		"truckers":       "Trucker",
		"material-sites": "MaterialSite",
	})
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.TradingPartnerType != "" {
		attributes["trading-partner-type"] = opts.TradingPartnerType
	}

	relationships := map[string]any{
		"organization": map[string]any{
			"data": map[string]any{
				"type": organizationType,
				"id":   organizationID,
			},
		},
		"partner": map[string]any{
			"data": map[string]any{
				"type": partnerType,
				"id":   partnerID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "trading-partners",
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

	body, _, err := client.Post(cmd.Context(), "/v1/trading-partners", jsonBody)
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

	row := buildTradingPartnerRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created trading partner %s\n", row.ID)
	return nil
}

func parseDoTradingPartnersCreateOptions(cmd *cobra.Command) (doTradingPartnersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	organization, _ := cmd.Flags().GetString("organization")
	partner, _ := cmd.Flags().GetString("partner")
	tradingPartnerType, _ := cmd.Flags().GetString("trading-partner-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTradingPartnersCreateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		Organization:       organization,
		Partner:            partner,
		TradingPartnerType: tradingPartnerType,
	}, nil
}

func parseTradingPartnerReference(value, label string, allowed map[string]string) (string, string, error) {
	parts := strings.SplitN(value, "|", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid %s format: %q (expected Type|ID, e.g. Broker|123)", label, value)
	}
	resourceType := strings.TrimSpace(parts[0])
	resourceID := strings.TrimSpace(parts[1])
	if resourceType == "" || resourceID == "" {
		return "", "", fmt.Errorf("invalid %s format: %q (expected Type|ID, e.g. Broker|123)", label, value)
	}

	jsonAPIType := normalizeTradingPartnerType(resourceType)
	if jsonAPIType == "" {
		return "", "", fmt.Errorf("unsupported %s type %q", label, resourceType)
	}
	if _, ok := allowed[jsonAPIType]; !ok {
		allowedLabels := make([]string, 0, len(allowed))
		for _, labelValue := range allowed {
			allowedLabels = append(allowedLabels, labelValue)
		}
		return "", "", fmt.Errorf("unsupported %s type %q (expected %s)", label, resourceType, strings.Join(allowedLabels, " or "))
	}

	return jsonAPIType, resourceID, nil
}

func normalizeTradingPartnerType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "broker", "brokers":
		return "brokers"
	case "customer", "customers":
		return "customers"
	case "trucker", "truckers":
		return "truckers"
	case "materialsite", "material-site", "material_site", "material-sites", "materialsites":
		return "material-sites"
	default:
		return ""
	}
}

func buildTradingPartnerRowFromSingle(resp jsonAPISingleResponse) tradingPartnerRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	row := tradingPartnerRow{
		ID:                 resp.Data.ID,
		TradingPartnerType: stringAttr(resp.Data.Attributes, "trading-partner-type"),
	}

	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationType = rel.Data.Type
		row.OrganizationID = rel.Data.ID
		row.OrganizationName = tradingPartnerNameFromIncluded(rel.Data, included)
	}

	if rel, ok := resp.Data.Relationships["partner"]; ok && rel.Data != nil {
		row.PartnerType = rel.Data.Type
		row.PartnerID = rel.Data.ID
		row.PartnerName = tradingPartnerNameFromIncluded(rel.Data, included)
	}

	return row
}
