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

type jobSitesListOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	NoAuth                      bool
	Limit                       int
	Offset                      int
	Name                        string
	NameLike                    string
	Active                      bool
	Broker                      string
	BrokerID                    string
	Customer                    string
	Q                           string
	MaterialSite                string
	HasMaterialSite             string
	IsStockpiling               string
	ActiveSince                 string
	AddressNear                 string
	ExternalIdentificationValue string
	ExternalJobNumber           string
}

type jobSiteRow struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Active     bool   `json:"is_active"`
	Customer   string `json:"customer,omitempty"`
	CustomerID string `json:"customer_id,omitempty"`
	Broker     string `json:"broker,omitempty"`
	BrokerID   string `json:"broker_id,omitempty"`
}

func newJobSitesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job sites",
		Long: `List job sites with filtering and pagination.

Returns a list of job sites. Use this to look up job site IDs for filtering
job production plans.

Output Columns:
  ID        Job site identifier (use this for --job-site filter)
  NAME      Job site name
  ACTIVE    Whether the site is active
  CUSTOMER  Customer name
  BROKER    Broker name`,
		Example: `  # List job sites
  xbe view job-sites list

  # Search by name
  xbe view job-sites list --name "Main Street"

  # List active job sites only
  xbe view job-sites list --active

  # Output as JSON
  xbe view job-sites list --json`,
		RunE: runJobSitesList,
	}
	initJobSitesListFlags(cmd)
	return cmd
}

func init() {
	jobSitesCmd.AddCommand(newJobSitesListCmd())
}

func initJobSitesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Bool("active", false, "Show only active job sites")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by name (exact match)")
	cmd.Flags().String("name-like", "", "Filter by name (partial/fuzzy match)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("customer", "", "Filter by customer ID (comma-separated for multiple)")
	cmd.Flags().String("q", "", "Full-text search")
	cmd.Flags().String("material-site", "", "Filter by material site ID (comma-separated for multiple)")
	cmd.Flags().String("has-material-site", "", "Filter by whether site has material site (true/false)")
	cmd.Flags().String("is-stockpiling", "", "Filter by stockpiling status (true/false)")
	cmd.Flags().String("active-since", "", "Filter by activity since date (YYYY-MM-DD)")
	cmd.Flags().String("address-near", "", "Filter by proximity to address (lat,lng,radius_miles)")
	cmd.Flags().String("broker-id", "", "Filter by broker ID (uses broker_id filter)")
	cmd.Flags().String("external-identification-value", "", "Filter by external identification value")
	cmd.Flags().String("external-job-number", "", "Filter by external job number")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobSitesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobSitesListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("sort", "name")
	query.Set("fields[job-sites]", "name,is-active,customer,broker")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "customer,broker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Name != "" {
		query.Set("filter[name]", opts.Name)
	}
	setFilterIfPresent(query, "filter[name-like]", opts.NameLike)
	if opts.Active {
		query.Set("filter[is-active]", "true")
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[broker-id]", opts.BrokerID)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[q]", opts.Q)
	setFilterIfPresent(query, "filter[material-site]", opts.MaterialSite)
	setFilterIfPresent(query, "filter[has-material-site]", opts.HasMaterialSite)
	setFilterIfPresent(query, "filter[is-stockpiling]", opts.IsStockpiling)
	setFilterIfPresent(query, "filter[active-since]", opts.ActiveSince)
	setFilterIfPresent(query, "filter[address-near]", opts.AddressNear)
	setFilterIfPresent(query, "filter[external-identification-value]", opts.ExternalIdentificationValue)
	setFilterIfPresent(query, "filter[external-job-number]", opts.ExternalJobNumber)

	body, _, err := client.Get(cmd.Context(), "/v1/job-sites", query)
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

	rows := buildJobSiteRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobSitesTable(cmd, rows)
}

func parseJobSitesListOptions(cmd *cobra.Command) (jobSitesListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return jobSitesListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return jobSitesListOptions{}, err
	}
	active, err := cmd.Flags().GetBool("active")
	if err != nil {
		return jobSitesListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return jobSitesListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return jobSitesListOptions{}, err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return jobSitesListOptions{}, err
	}
	nameLike, err := cmd.Flags().GetString("name-like")
	if err != nil {
		return jobSitesListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return jobSitesListOptions{}, err
	}
	brokerID, err := cmd.Flags().GetString("broker-id")
	if err != nil {
		return jobSitesListOptions{}, err
	}
	customer, err := cmd.Flags().GetString("customer")
	if err != nil {
		return jobSitesListOptions{}, err
	}
	q, err := cmd.Flags().GetString("q")
	if err != nil {
		return jobSitesListOptions{}, err
	}
	materialSite, err := cmd.Flags().GetString("material-site")
	if err != nil {
		return jobSitesListOptions{}, err
	}
	hasMaterialSite, err := cmd.Flags().GetString("has-material-site")
	if err != nil {
		return jobSitesListOptions{}, err
	}
	isStockpiling, err := cmd.Flags().GetString("is-stockpiling")
	if err != nil {
		return jobSitesListOptions{}, err
	}
	activeSince, err := cmd.Flags().GetString("active-since")
	if err != nil {
		return jobSitesListOptions{}, err
	}
	addressNear, err := cmd.Flags().GetString("address-near")
	if err != nil {
		return jobSitesListOptions{}, err
	}
	externalIdentificationValue, err := cmd.Flags().GetString("external-identification-value")
	if err != nil {
		return jobSitesListOptions{}, err
	}
	externalJobNumber, err := cmd.Flags().GetString("external-job-number")
	if err != nil {
		return jobSitesListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return jobSitesListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return jobSitesListOptions{}, err
	}

	return jobSitesListOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		NoAuth:                      noAuth,
		Active:                      active,
		Limit:                       limit,
		Offset:                      offset,
		Name:                        name,
		NameLike:                    nameLike,
		Broker:                      broker,
		BrokerID:                    brokerID,
		Customer:                    customer,
		Q:                           q,
		MaterialSite:                materialSite,
		HasMaterialSite:             hasMaterialSite,
		IsStockpiling:               isStockpiling,
		ActiveSince:                 activeSince,
		AddressNear:                 addressNear,
		ExternalIdentificationValue: externalIdentificationValue,
		ExternalJobNumber:           externalJobNumber,
	}, nil
}

func buildJobSiteRows(resp jsonAPIResponse) []jobSiteRow {
	// Build included lookup
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]jobSiteRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := jobSiteRow{
			ID:     resource.ID,
			Name:   stringAttr(resource.Attributes, "name"),
			Active: boolAttr(resource.Attributes, "is-active"),
		}

		// Resolve customer
		if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
			row.CustomerID = rel.Data.ID
			if customer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.Customer = stringAttr(customer.Attributes, "company-name")
			}
		}

		// Resolve broker
		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.Broker = stringAttr(broker.Attributes, "company-name")
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderJobSitesTable(cmd *cobra.Command, rows []jobSiteRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job sites found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tACTIVE\tCUSTOMER\tBROKER")
	for _, row := range rows {
		activeStr := ""
		if row.Active {
			activeStr = "Yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 40),
			activeStr,
			truncateString(row.Customer, 25),
			truncateString(row.Broker, 25),
		)
	}
	return writer.Flush()
}
