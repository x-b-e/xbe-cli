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

type tradingPartnersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type tradingPartnerDetails struct {
	ID                        string   `json:"id"`
	TradingPartnerType        string   `json:"trading_partner_type,omitempty"`
	OrganizationType          string   `json:"organization_type,omitempty"`
	OrganizationID            string   `json:"organization_id,omitempty"`
	OrganizationName          string   `json:"organization_name,omitempty"`
	PartnerType               string   `json:"partner_type,omitempty"`
	PartnerID                 string   `json:"partner_id,omitempty"`
	PartnerName               string   `json:"partner_name,omitempty"`
	CreatedAt                 string   `json:"created_at,omitempty"`
	UpdatedAt                 string   `json:"updated_at,omitempty"`
	ExternalIdentificationIDs []string `json:"external_identification_ids,omitempty"`
}

func newTradingPartnersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show a trading partner",
		Long: `Show full trading-partner details by ID.

Includes organization and partner relationships and any external identification links.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a trading partner
  xbe view trading-partners show 123`,
		Args: cobra.ExactArgs(1),
		RunE: runTradingPartnersShow,
	}
	initTradingPartnersShowFlags(cmd)
	return cmd
}

func init() {
	tradingPartnersCmd.AddCommand(newTradingPartnersShowCmd())
}

func initTradingPartnersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTradingPartnersShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseTradingPartnersShowOptions(cmd)
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
	query.Set("fields[trading-partners]", "trading-partner-type,created-at,updated-at,organization,partner,external-identifications")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[material-sites]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/trading-partners/"+args[0], query)
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

	details := buildTradingPartnerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTradingPartnerDetails(cmd, details)
}

func parseTradingPartnersShowOptions(cmd *cobra.Command) (tradingPartnersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tradingPartnersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTradingPartnerDetails(resp jsonAPISingleResponse) tradingPartnerDetails {
	attrs := resp.Data.Attributes
	details := tradingPartnerDetails{
		ID:                 resp.Data.ID,
		TradingPartnerType: stringAttr(attrs, "trading-partner-type"),
		CreatedAt:          formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:          formatDateTime(stringAttr(attrs, "updated-at")),
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationType = rel.Data.Type
		details.OrganizationID = rel.Data.ID
		details.OrganizationName = tradingPartnerNameFromIncluded(rel.Data, included)
	}

	if rel, ok := resp.Data.Relationships["partner"]; ok && rel.Data != nil {
		details.PartnerType = rel.Data.Type
		details.PartnerID = rel.Data.ID
		details.PartnerName = tradingPartnerNameFromIncluded(rel.Data, included)
	}

	if rel, ok := resp.Data.Relationships["external-identifications"]; ok {
		details.ExternalIdentificationIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderTradingPartnerDetails(cmd *cobra.Command, details tradingPartnerDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TradingPartnerType != "" {
		fmt.Fprintf(out, "Trading Partner Type: %s\n", details.TradingPartnerType)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Organization:")
	if details.OrganizationType != "" {
		fmt.Fprintf(out, "  Type: %s\n", details.OrganizationType)
	}
	if details.OrganizationID != "" {
		fmt.Fprintf(out, "  ID: %s\n", details.OrganizationID)
	}
	if details.OrganizationName != "" {
		fmt.Fprintf(out, "  Name: %s\n", details.OrganizationName)
	}
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Partner:")
	if details.PartnerType != "" {
		fmt.Fprintf(out, "  Type: %s\n", details.PartnerType)
	}
	if details.PartnerID != "" {
		fmt.Fprintf(out, "  ID: %s\n", details.PartnerID)
	}
	if details.PartnerName != "" {
		fmt.Fprintf(out, "  Name: %s\n", details.PartnerName)
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
