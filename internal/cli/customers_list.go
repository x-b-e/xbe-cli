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

type customersListOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	NoAuth                      bool
	Limit                       int
	Offset                      int
	Name                        string
	IsActive                    bool
	Broker                      string
	Favorite                    string
	IsControlledByBroker        string
	TrailerClassification       string
	IsOnlyForEquipmentMovement  string
	BrokerCustomerID            string
	ExternalIdentificationValue string
}

func newCustomersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List customers",
		Long: `List customers with filtering and pagination.

Returns a list of customers matching the specified criteria.

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filtering:
  Multiple filters can be combined. All filters use AND logic.

Use Case:
  Find customer IDs for filtering posts by creator:
    xbe view posts list --creator "Customer|<id>"`,
		Example: `  # List customers
  xbe view customers list

  # Search by company name
  xbe view customers list --name "Acme"

  # Filter by active status
  xbe view customers list --active

  # Paginate results
  xbe view customers list --limit 20 --offset 40

  # Output as JSON
  xbe view customers list --json`,
		RunE: runCustomersList,
	}
	initCustomersListFlags(cmd)
	return cmd
}

func init() {
	customersCmd.AddCommand(newCustomersListCmd())
}

func initCustomersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by company name (partial match)")
	cmd.Flags().Bool("active", false, "Filter to only active customers")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("favorite", "", "Filter by favorite status (true/false)")
	cmd.Flags().String("is-controlled-by-broker", "", "Filter by broker control status (true/false)")
	cmd.Flags().String("trailer-classification", "", "Filter by trailer classification")
	cmd.Flags().String("is-only-for-equipment-movement", "", "Filter by equipment movement only status (true/false)")
	cmd.Flags().String("broker-customer-id", "", "Filter by broker-assigned customer ID")
	cmd.Flags().String("external-identification-value", "", "Filter by external identification value")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCustomersListOptions(cmd)
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
	query.Set("fields[customers]", "company-name,is-active,broker")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "broker")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[company-name]", opts.Name)
	if opts.IsActive {
		query.Set("filter[is_active]", "true")
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[favorite]", opts.Favorite)
	setFilterIfPresent(query, "filter[is-controlled-by-broker]", opts.IsControlledByBroker)
	setFilterIfPresent(query, "filter[trailer-classification]", opts.TrailerClassification)
	setFilterIfPresent(query, "filter[is-only-for-equipment-movement]", opts.IsOnlyForEquipmentMovement)
	setFilterIfPresent(query, "filter[broker-customer-id]", opts.BrokerCustomerID)
	setFilterIfPresent(query, "filter[external-identification-value]", opts.ExternalIdentificationValue)

	body, _, err := client.Get(cmd.Context(), "/v1/customers", query)
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

	if opts.JSON {
		rows := buildCustomerRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCustomersList(cmd, resp)
}

func parseCustomersListOptions(cmd *cobra.Command) (customersListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return customersListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return customersListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return customersListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return customersListOptions{}, err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return customersListOptions{}, err
	}
	isActive, err := cmd.Flags().GetBool("active")
	if err != nil {
		return customersListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return customersListOptions{}, err
	}
	favorite, err := cmd.Flags().GetString("favorite")
	if err != nil {
		return customersListOptions{}, err
	}
	isControlledByBroker, err := cmd.Flags().GetString("is-controlled-by-broker")
	if err != nil {
		return customersListOptions{}, err
	}
	trailerClassification, err := cmd.Flags().GetString("trailer-classification")
	if err != nil {
		return customersListOptions{}, err
	}
	isOnlyForEquipmentMovement, err := cmd.Flags().GetString("is-only-for-equipment-movement")
	if err != nil {
		return customersListOptions{}, err
	}
	brokerCustomerID, err := cmd.Flags().GetString("broker-customer-id")
	if err != nil {
		return customersListOptions{}, err
	}
	externalIdentificationValue, err := cmd.Flags().GetString("external-identification-value")
	if err != nil {
		return customersListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return customersListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return customersListOptions{}, err
	}

	return customersListOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		NoAuth:                      noAuth,
		Limit:                       limit,
		Offset:                      offset,
		Name:                        name,
		IsActive:                    isActive,
		Broker:                      broker,
		Favorite:                    favorite,
		IsControlledByBroker:        isControlledByBroker,
		TrailerClassification:       trailerClassification,
		IsOnlyForEquipmentMovement:  isOnlyForEquipmentMovement,
		BrokerCustomerID:            brokerCustomerID,
		ExternalIdentificationValue: externalIdentificationValue,
	}, nil
}

type customerRow struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Broker   string `json:"broker"`
	IsActive bool   `json:"is_active"`
}

func buildCustomerRows(resp jsonAPIResponse) []customerRow {
	// Build included map for broker lookup
	included := make(map[string]map[string]any)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc.Attributes
	}

	rows := make([]customerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		brokerName := ""
		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if attrs, ok := included[key]; ok {
				brokerName = strings.TrimSpace(stringAttr(attrs, "company-name"))
			}
		}

		rows = append(rows, customerRow{
			ID:       resource.ID,
			Name:     strings.TrimSpace(stringAttr(resource.Attributes, "company-name")),
			Broker:   brokerName,
			IsActive: boolAttr(resource.Attributes, "is-active"),
		})
	}
	return rows
}

func renderCustomersList(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildCustomerRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No customers found.")
		return nil
	}

	const nameMax = 50
	const brokerMax = 30

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tBROKER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, nameMax),
			truncateString(row.Broker, brokerMax),
		)
	}
	return writer.Flush()
}
