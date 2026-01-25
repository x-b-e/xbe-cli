package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type customerVendorsListOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	NoAuth                      bool
	Limit                       int
	Offset                      int
	Sort                        string
	Customer                    string
	Organization                string
	OrganizationType            string
	OrganizationID              string
	NotOrganizationType         string
	Partner                     string
	PartnerType                 string
	PartnerID                   string
	NotPartnerType              string
	TradingPartnerType          string
	ExternalIdentificationValue string
	CreatedAtMin                string
	CreatedAtMax                string
	UpdatedAtMin                string
	UpdatedAtMax                string
}

type customerVendorRow struct {
	ID                                 string `json:"id"`
	CustomerID                         string `json:"customer_id,omitempty"`
	CustomerName                       string `json:"customer_name,omitempty"`
	VendorID                           string `json:"vendor_id,omitempty"`
	VendorName                         string `json:"vendor_name,omitempty"`
	ExternalAccountingCustomerVendorID string `json:"external_accounting_customer_vendor_id,omitempty"`
	TradingPartnerType                 string `json:"trading_partner_type,omitempty"`
}

func newCustomerVendorsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List customer vendors",
		Long: `List customer vendors with filtering and pagination.

Customer vendors represent trading partner links between a customer and a vendor
(trucker).

Output Columns:
  ID           Customer-vendor relationship ID
  CUSTOMER     Customer name (falls back to ID)
  VENDOR       Vendor name (falls back to ID)
  EXT ACCT ID  External accounting customer vendor ID

Filters:
  --customer                              Filter by customer ID
  --organization                          Filter by organization (Type|ID, e.g. Customer|123)
  --organization-type                     Filter by organization type
  --organization-id                       Filter by organization ID
  --not-organization-type                 Exclude by organization type
  --partner                               Filter by partner (Type|ID, e.g. Trucker|456)
  --partner-type                          Filter by partner type
  --partner-id                            Filter by partner ID
  --not-partner-type                      Exclude by partner type
  --trading-partner-type                  Filter by trading partner type (e.g. CustomerVendor)
  --external-identification-value         Filter by external identification value
  --created-at-min                        Filter by created-at on/after (ISO 8601)
  --created-at-max                        Filter by created-at on/before (ISO 8601)
  --updated-at-min                        Filter by updated-at on/after (ISO 8601)
  --updated-at-max                        Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List customer vendors
  xbe view customer-vendors list

  # Filter by customer
  xbe view customer-vendors list --customer 123

  # Filter by customer and vendor via polymorphic filters
  xbe view customer-vendors list --organization "Customer|123" --partner "Trucker|456"

  # Output as JSON
  xbe view customer-vendors list --json`,
		Args: cobra.NoArgs,
		RunE: runCustomerVendorsList,
	}
	initCustomerVendorsListFlags(cmd)
	return cmd
}

func init() {
	customerVendorsCmd.AddCommand(newCustomerVendorsListCmd())
}

func initCustomerVendorsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("organization", "", "Filter by organization (Type|ID, e.g. Customer|123)")
	cmd.Flags().String("organization-type", "", "Filter by organization type")
	cmd.Flags().String("organization-id", "", "Filter by organization ID (use with --organization-type)")
	cmd.Flags().String("not-organization-type", "", "Exclude by organization type")
	cmd.Flags().String("partner", "", "Filter by partner (Type|ID, e.g. Trucker|456)")
	cmd.Flags().String("partner-type", "", "Filter by partner type")
	cmd.Flags().String("partner-id", "", "Filter by partner ID (use with --partner-type)")
	cmd.Flags().String("not-partner-type", "", "Exclude by partner type")
	cmd.Flags().String("trading-partner-type", "", "Filter by trading partner type (e.g. CustomerVendor)")
	cmd.Flags().String("external-identification-value", "", "Filter by external identification value")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomerVendorsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCustomerVendorsListOptions(cmd)
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
	query.Set("include", "organization,partner")
	query.Set("fields[customer-vendors]", "external-accounting-customer-vendor-id,trading-partner-type,organization,partner")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[customers]", "company-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	organizationFilter := strings.TrimSpace(opts.Organization)
	if organizationFilter == "" {
		if opts.OrganizationType != "" && opts.OrganizationID != "" {
			organizationFilter = opts.OrganizationType + "|" + opts.OrganizationID
		} else if opts.Customer != "" {
			organizationFilter = "Customer|" + opts.Customer
		}
	}
	if organizationFilter != "" {
		query.Set("filter[organization]", organizationFilter)
		if opts.OrganizationID != "" {
			query.Set("filter[organization-id]", organizationFilter)
		}
	} else if opts.OrganizationID != "" {
		return fmt.Errorf("--organization-id requires --organization-type or --organization")
	}
	setFilterIfPresent(query, "filter[organization-type]", opts.OrganizationType)
	setFilterIfPresent(query, "filter[not-organization-type]", opts.NotOrganizationType)

	partnerFilter := strings.TrimSpace(opts.Partner)
	if partnerFilter == "" && opts.PartnerType != "" && opts.PartnerID != "" {
		partnerFilter = opts.PartnerType + "|" + opts.PartnerID
	}
	if partnerFilter != "" {
		query.Set("filter[partner]", partnerFilter)
		if opts.PartnerID != "" {
			query.Set("filter[partner-id]", partnerFilter)
		}
	} else if opts.PartnerID != "" {
		return fmt.Errorf("--partner-id requires --partner-type or --partner")
	}
	setFilterIfPresent(query, "filter[partner-type]", opts.PartnerType)
	setFilterIfPresent(query, "filter[not-partner-type]", opts.NotPartnerType)
	setFilterIfPresent(query, "filter[trading-partner-type]", opts.TradingPartnerType)
	setFilterIfPresent(query, "filter[external-identification-value]", opts.ExternalIdentificationValue)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/customer-vendors", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildCustomerVendorRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCustomerVendorsTable(cmd, rows)
}

func parseCustomerVendorsListOptions(cmd *cobra.Command) (customerVendorsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	customer, _ := cmd.Flags().GetString("customer")
	organization, _ := cmd.Flags().GetString("organization")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	notOrganizationType, _ := cmd.Flags().GetString("not-organization-type")
	partner, _ := cmd.Flags().GetString("partner")
	partnerType, _ := cmd.Flags().GetString("partner-type")
	partnerID, _ := cmd.Flags().GetString("partner-id")
	notPartnerType, _ := cmd.Flags().GetString("not-partner-type")
	tradingPartnerType, _ := cmd.Flags().GetString("trading-partner-type")
	externalIdentificationValue, _ := cmd.Flags().GetString("external-identification-value")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return customerVendorsListOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		NoAuth:                      noAuth,
		Limit:                       limit,
		Offset:                      offset,
		Sort:                        sort,
		Customer:                    customer,
		Organization:                organization,
		OrganizationType:            organizationType,
		OrganizationID:              organizationID,
		NotOrganizationType:         notOrganizationType,
		Partner:                     partner,
		PartnerType:                 partnerType,
		PartnerID:                   partnerID,
		NotPartnerType:              notPartnerType,
		TradingPartnerType:          tradingPartnerType,
		ExternalIdentificationValue: externalIdentificationValue,
		CreatedAtMin:                createdAtMin,
		CreatedAtMax:                createdAtMax,
		UpdatedAtMin:                updatedAtMin,
		UpdatedAtMax:                updatedAtMax,
	}, nil
}

func buildCustomerVendorRows(resp jsonAPIResponse) []customerVendorRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]customerVendorRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := customerVendorRow{
			ID:                                 resource.ID,
			ExternalAccountingCustomerVendorID: stringAttr(resource.Attributes, "external-accounting-customer-vendor-id"),
			TradingPartnerType:                 stringAttr(resource.Attributes, "trading-partner-type"),
		}

		if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
			row.CustomerID = rel.Data.ID
			row.CustomerName = customerVendorNameFromIncluded(rel.Data, included)
		} else if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			row.CustomerID = rel.Data.ID
			row.CustomerName = customerVendorNameFromIncluded(rel.Data, included)
		}

		if rel, ok := resource.Relationships["vendor"]; ok && rel.Data != nil {
			row.VendorID = rel.Data.ID
			row.VendorName = customerVendorNameFromIncluded(rel.Data, included)
		} else if rel, ok := resource.Relationships["partner"]; ok && rel.Data != nil {
			row.VendorID = rel.Data.ID
			row.VendorName = customerVendorNameFromIncluded(rel.Data, included)
		}

		rows = append(rows, row)
	}

	return rows
}

func customerVendorNameFromIncluded(rel *jsonAPIResourceIdentifier, included map[string]jsonAPIResource) string {
	if rel == nil {
		return ""
	}
	key := resourceKey(rel.Type, rel.ID)
	if inc, ok := included[key]; ok {
		return firstNonEmpty(
			strings.TrimSpace(stringAttr(inc.Attributes, "company-name")),
			strings.TrimSpace(stringAttr(inc.Attributes, "name")),
		)
	}
	return ""
}

func renderCustomerVendorsTable(cmd *cobra.Command, rows []customerVendorRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No customer vendors found.")
		return nil
	}

	const customerMax = 24
	const vendorMax = 24

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCUSTOMER\tVENDOR\tEXT ACCT ID")
	for _, row := range rows {
		customerLabel := row.CustomerName
		if customerLabel == "" {
			customerLabel = row.CustomerID
		}
		vendorLabel := row.VendorName
		if vendorLabel == "" {
			vendorLabel = row.VendorID
		}
		extID := row.ExternalAccountingCustomerVendorID
		if extID == "" {
			extID = "-"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(customerLabel, customerMax),
			truncateString(vendorLabel, vendorMax),
			extID,
		)
	}
	return writer.Flush()
}
