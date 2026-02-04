package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type brokerVendorsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type brokerVendorDetails struct {
	ID                               string   `json:"id"`
	TradingPartnerType               string   `json:"trading_partner_type,omitempty"`
	ExternalAccountingBrokerVendorID string   `json:"external_accounting_broker_vendor_id,omitempty"`
	BrokerID                         string   `json:"broker_id,omitempty"`
	BrokerName                       string   `json:"broker_name,omitempty"`
	VendorID                         string   `json:"vendor_id,omitempty"`
	VendorName                       string   `json:"vendor_name,omitempty"`
	CreatedAt                        string   `json:"created_at,omitempty"`
	UpdatedAt                        string   `json:"updated_at,omitempty"`
	ExternalIdentificationIDs        []string `json:"external_identification_ids,omitempty"`
}

func newBrokerVendorsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show a broker vendor",
		Long: `Show full broker-vendor details by ID.

Includes the broker and vendor relationships and any external identification links.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a broker-vendor relationship
  xbe view broker-vendors show 123`,
		Args: cobra.ExactArgs(1),
		RunE: runBrokerVendorsShow,
	}
	initBrokerVendorsShowFlags(cmd)
	return cmd
}

func init() {
	brokerVendorsCmd.AddCommand(newBrokerVendorsShowCmd())
}

func initBrokerVendorsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokerVendorsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseBrokerVendorsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "organization,partner,external-identifications")
	query.Set("fields[broker-vendors]", "external-accounting-broker-vendor-id,trading-partner-type,created-at,updated-at,organization,partner,external-identifications")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[material-sites]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/broker-vendors/"+args[0], query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildBrokerVendorDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderBrokerVendorDetails(cmd, details)
}

func parseBrokerVendorsShowOptions(cmd *cobra.Command) (brokerVendorsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return brokerVendorsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildBrokerVendorDetails(resp jsonAPISingleResponse) brokerVendorDetails {
	attrs := resp.Data.Attributes
	details := brokerVendorDetails{
		ID:                               resp.Data.ID,
		TradingPartnerType:               stringAttr(attrs, "trading-partner-type"),
		ExternalAccountingBrokerVendorID: stringAttr(attrs, "external-accounting-broker-vendor-id"),
		CreatedAt:                        formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:                        formatDateTime(stringAttr(attrs, "updated-at")),
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		details.BrokerName = brokerVendorNameFromIncluded(rel.Data, included)
	} else if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		details.BrokerName = brokerVendorNameFromIncluded(rel.Data, included)
	}

	if rel, ok := resp.Data.Relationships["vendor"]; ok && rel.Data != nil {
		details.VendorID = rel.Data.ID
		details.VendorName = brokerVendorNameFromIncluded(rel.Data, included)
	} else if rel, ok := resp.Data.Relationships["partner"]; ok && rel.Data != nil {
		details.VendorID = rel.Data.ID
		details.VendorName = brokerVendorNameFromIncluded(rel.Data, included)
	}

	if rel, ok := resp.Data.Relationships["external-identifications"]; ok {
		details.ExternalIdentificationIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderBrokerVendorDetails(cmd *cobra.Command, details brokerVendorDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TradingPartnerType != "" {
		fmt.Fprintf(out, "Trading Partner Type: %s\n", details.TradingPartnerType)
	}
	if details.ExternalAccountingBrokerVendorID != "" {
		fmt.Fprintf(out, "External Accounting Broker Vendor ID: %s\n", details.ExternalAccountingBrokerVendorID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Broker:")
	if details.BrokerID != "" {
		fmt.Fprintf(out, "  ID: %s\n", details.BrokerID)
	}
	if details.BrokerName != "" {
		fmt.Fprintf(out, "  Name: %s\n", details.BrokerName)
	}
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Vendor:")
	if details.VendorID != "" {
		fmt.Fprintf(out, "  ID: %s\n", details.VendorID)
	}
	if details.VendorName != "" {
		fmt.Fprintf(out, "  Name: %s\n", details.VendorName)
	}
	fmt.Fprintln(out, "")

	if len(details.ExternalIdentificationIDs) > 0 {
		fmt.Fprintln(out, "External Identifications:")
		for _, id := range details.ExternalIdentificationIDs {
			fmt.Fprintf(out, "  - %s\n", id)
		}
	}

	return nil
}
