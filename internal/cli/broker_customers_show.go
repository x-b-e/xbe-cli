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

type brokerCustomersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type brokerCustomerDetails struct {
	ID                                 string   `json:"id"`
	TradingPartnerType                 string   `json:"trading_partner_type,omitempty"`
	ExternalAccountingBrokerCustomerID string   `json:"external_accounting_broker_customer_id,omitempty"`
	BrokerID                           string   `json:"broker_id,omitempty"`
	BrokerName                         string   `json:"broker_name,omitempty"`
	CustomerID                         string   `json:"customer_id,omitempty"`
	CustomerName                       string   `json:"customer_name,omitempty"`
	CreatedAt                          string   `json:"created_at,omitempty"`
	UpdatedAt                          string   `json:"updated_at,omitempty"`
	ExternalIdentificationIDs          []string `json:"external_identification_ids,omitempty"`
}

func newBrokerCustomersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show a broker customer",
		Long: `Show full broker-customer details by ID.

Includes the broker and customer relationships and any external identification links.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a broker-customer relationship
  xbe view broker-customers show 123`,
		Args: cobra.ExactArgs(1),
		RunE: runBrokerCustomersShow,
	}
	initBrokerCustomersShowFlags(cmd)
	return cmd
}

func init() {
	brokerCustomersCmd.AddCommand(newBrokerCustomersShowCmd())
}

func initBrokerCustomersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokerCustomersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseBrokerCustomersShowOptions(cmd)
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
	query.Set("fields[broker-customers]", "external-accounting-broker-customer-id,trading-partner-type,created-at,updated-at,organization,partner,external-identifications")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")

	body, _, err := client.Get(cmd.Context(), "/v1/broker-customers/"+args[0], query)
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

	details := buildBrokerCustomerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderBrokerCustomerDetails(cmd, details)
}

func parseBrokerCustomersShowOptions(cmd *cobra.Command) (brokerCustomersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return brokerCustomersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildBrokerCustomerDetails(resp jsonAPISingleResponse) brokerCustomerDetails {
	attrs := resp.Data.Attributes
	details := brokerCustomerDetails{
		ID:                                 resp.Data.ID,
		TradingPartnerType:                 stringAttr(attrs, "trading-partner-type"),
		ExternalAccountingBrokerCustomerID: stringAttr(attrs, "external-accounting-broker-customer-id"),
		CreatedAt:                          formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:                          formatDateTime(stringAttr(attrs, "updated-at")),
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		details.BrokerName = brokerCustomerNameFromIncluded(rel.Data, included)
	} else if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		details.BrokerName = brokerCustomerNameFromIncluded(rel.Data, included)
	}

	if rel, ok := resp.Data.Relationships["customer"]; ok && rel.Data != nil {
		details.CustomerID = rel.Data.ID
		details.CustomerName = brokerCustomerNameFromIncluded(rel.Data, included)
	} else if rel, ok := resp.Data.Relationships["partner"]; ok && rel.Data != nil {
		details.CustomerID = rel.Data.ID
		details.CustomerName = brokerCustomerNameFromIncluded(rel.Data, included)
	}

	if rel, ok := resp.Data.Relationships["external-identifications"]; ok {
		details.ExternalIdentificationIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderBrokerCustomerDetails(cmd *cobra.Command, details brokerCustomerDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TradingPartnerType != "" {
		fmt.Fprintf(out, "Trading Partner Type: %s\n", details.TradingPartnerType)
	}
	if details.ExternalAccountingBrokerCustomerID != "" {
		fmt.Fprintf(out, "External Accounting Broker Customer ID: %s\n", details.ExternalAccountingBrokerCustomerID)
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

	fmt.Fprintln(out, "Customer:")
	if details.CustomerID != "" {
		fmt.Fprintf(out, "  ID: %s\n", details.CustomerID)
	}
	if details.CustomerName != "" {
		fmt.Fprintf(out, "  Name: %s\n", details.CustomerName)
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
