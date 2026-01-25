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

type incidentSubscriptionsListOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	NoAuth             bool
	Limit              int
	Offset             int
	Sort               string
	User               string
	Customer           string
	Broker             string
	MaterialSupplier   string
	Kind               string
	ContactMethod      string
	Incident           string
	IncidentStartOn    string
	IncidentStartOnMin string
	IncidentStartOnMax string
}

type incidentSubscriptionRow struct {
	ID                      string `json:"id"`
	UserID                  string `json:"user_id,omitempty"`
	Kind                    string `json:"kind,omitempty"`
	ContactMethod           string `json:"contact_method,omitempty"`
	CalculatedContactMethod string `json:"calculated_contact_method,omitempty"`
	OrganizationType        string `json:"organization_type,omitempty"`
	OrganizationID          string `json:"organization_id,omitempty"`
	CustomerID              string `json:"customer_id,omitempty"`
	BrokerID                string `json:"broker_id,omitempty"`
	MaterialSupplierID      string `json:"material_supplier_id,omitempty"`
	IncidentType            string `json:"incident_type,omitempty"`
	IncidentID              string `json:"incident_id,omitempty"`
}

func newIncidentSubscriptionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List incident subscriptions",
		Long: `List incident subscriptions with filtering and pagination.

Output Columns:
  ID            Subscription identifier
  USER          User ID
  KIND          Incident kind filter
  CONTACT       Contact method (explicit or calculated)
  ORGANIZATION  Organization scope (type/id)
  INCIDENT      Incident scope (type/id)

Filters:
  --user                  Filter by user ID
  --customer              Filter by customer ID
  --broker                Filter by broker ID
  --material-supplier     Filter by material supplier ID
  --kind                  Filter by incident kind
  --contact-method        Filter by contact method (email_address, mobile_number)
  --incident              Filter by incident ID
  --incident-start-on     Filter by incident start date
  --incident-start-on-min Filter by minimum incident start date
  --incident-start-on-max Filter by maximum incident start date

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List incident subscriptions
  xbe view incident-subscriptions list

  # Filter by broker and kind
  xbe view incident-subscriptions list --broker 123 --kind safety

  # Filter by contact method
  xbe view incident-subscriptions list --contact-method email_address

  # Output as JSON
  xbe view incident-subscriptions list --json`,
		Args: cobra.NoArgs,
		RunE: runIncidentSubscriptionsList,
	}
	initIncidentSubscriptionsListFlags(cmd)
	return cmd
}

func init() {
	incidentSubscriptionsCmd.AddCommand(newIncidentSubscriptionsListCmd())
}

func initIncidentSubscriptionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("material-supplier", "", "Filter by material supplier ID")
	cmd.Flags().String("kind", "", "Filter by incident kind")
	cmd.Flags().String("contact-method", "", "Filter by contact method (email_address, mobile_number)")
	cmd.Flags().String("incident", "", "Filter by incident ID")
	cmd.Flags().String("incident-start-on", "", "Filter by incident start date (YYYY-MM-DD)")
	cmd.Flags().String("incident-start-on-min", "", "Filter by minimum incident start date (YYYY-MM-DD)")
	cmd.Flags().String("incident-start-on-max", "", "Filter by maximum incident start date (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIncidentSubscriptionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseIncidentSubscriptionsListOptions(cmd)
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
	query.Set("fields[incident-subscriptions]", "kind,contact-method,calculated-contact-method,user,organization,customer,broker,material-supplier,incident")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[material_supplier]", opts.MaterialSupplier)
	setFilterIfPresent(query, "filter[kind]", opts.Kind)
	setFilterIfPresent(query, "filter[contact_method]", opts.ContactMethod)
	setFilterIfPresent(query, "filter[incident]", opts.Incident)
	setFilterIfPresent(query, "filter[incident_start_on]", opts.IncidentStartOn)
	setFilterIfPresent(query, "filter[incident_start_on_min]", opts.IncidentStartOnMin)
	setFilterIfPresent(query, "filter[incident_start_on_max]", opts.IncidentStartOnMax)

	body, _, err := client.Get(cmd.Context(), "/v1/incident-subscriptions", query)
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

	rows := buildIncidentSubscriptionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderIncidentSubscriptionsTable(cmd, rows)
}

func parseIncidentSubscriptionsListOptions(cmd *cobra.Command) (incidentSubscriptionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	user, _ := cmd.Flags().GetString("user")
	customer, _ := cmd.Flags().GetString("customer")
	broker, _ := cmd.Flags().GetString("broker")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
	kind, _ := cmd.Flags().GetString("kind")
	contactMethod, _ := cmd.Flags().GetString("contact-method")
	incident, _ := cmd.Flags().GetString("incident")
	incidentStartOn, _ := cmd.Flags().GetString("incident-start-on")
	incidentStartOnMin, _ := cmd.Flags().GetString("incident-start-on-min")
	incidentStartOnMax, _ := cmd.Flags().GetString("incident-start-on-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return incidentSubscriptionsListOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		NoAuth:             noAuth,
		Limit:              limit,
		Offset:             offset,
		Sort:               sort,
		User:               user,
		Customer:           customer,
		Broker:             broker,
		MaterialSupplier:   materialSupplier,
		Kind:               kind,
		ContactMethod:      contactMethod,
		Incident:           incident,
		IncidentStartOn:    incidentStartOn,
		IncidentStartOnMin: incidentStartOnMin,
		IncidentStartOnMax: incidentStartOnMax,
	}, nil
}

func buildIncidentSubscriptionRows(resp jsonAPIResponse) []incidentSubscriptionRow {
	rows := make([]incidentSubscriptionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildIncidentSubscriptionRow(resource))
	}
	return rows
}

func buildIncidentSubscriptionRow(resource jsonAPIResource) incidentSubscriptionRow {
	attrs := resource.Attributes
	row := incidentSubscriptionRow{
		ID:                      resource.ID,
		Kind:                    stringAttr(attrs, "kind"),
		ContactMethod:           stringAttr(attrs, "contact-method"),
		CalculatedContactMethod: stringAttr(attrs, "calculated-contact-method"),
	}

	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationType = rel.Data.Type
		row.OrganizationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
		row.CustomerID = rel.Data.ID
		if row.OrganizationType == "" {
			row.OrganizationType = rel.Data.Type
			row.OrganizationID = rel.Data.ID
		}
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		if row.OrganizationType == "" {
			row.OrganizationType = rel.Data.Type
			row.OrganizationID = rel.Data.ID
		}
	}
	if rel, ok := resource.Relationships["material-supplier"]; ok && rel.Data != nil {
		row.MaterialSupplierID = rel.Data.ID
		if row.OrganizationType == "" {
			row.OrganizationType = rel.Data.Type
			row.OrganizationID = rel.Data.ID
		}
	}
	if rel, ok := resource.Relationships["incident"]; ok && rel.Data != nil {
		row.IncidentType = rel.Data.Type
		row.IncidentID = rel.Data.ID
	}

	return row
}

func buildIncidentSubscriptionRowFromSingle(resp jsonAPISingleResponse) incidentSubscriptionRow {
	return buildIncidentSubscriptionRow(resp.Data)
}

func renderIncidentSubscriptionsTable(cmd *cobra.Command, rows []incidentSubscriptionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No incident subscriptions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tUSER\tKIND\tCONTACT\tORGANIZATION\tINCIDENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.UserID,
			truncateString(row.Kind, 18),
			truncateString(incidentSubscriptionContactDisplay(row), 20),
			incidentSubscriptionRelationshipDisplay(row.OrganizationType, row.OrganizationID),
			incidentSubscriptionRelationshipDisplay(row.IncidentType, row.IncidentID),
		)
	}
	return writer.Flush()
}

func incidentSubscriptionContactDisplay(row incidentSubscriptionRow) string {
	if row.ContactMethod != "" {
		return row.ContactMethod
	}
	return row.CalculatedContactMethod
}

func incidentSubscriptionRelationshipDisplay(typ, id string) string {
	if typ == "" || id == "" {
		return ""
	}
	return typ + "/" + id
}
