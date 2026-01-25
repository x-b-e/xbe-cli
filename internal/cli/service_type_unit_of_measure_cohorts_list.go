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

type serviceTypeUnitOfMeasureCohortsListOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	NoAuth                     bool
	Limit                      int
	Offset                     int
	Sort                       string
	Customer                   string
	ServiceTypeUnitOfMeasureID string
}

type serviceTypeUnitOfMeasureCohortRow struct {
	ID                          string   `json:"id"`
	Name                        string   `json:"name,omitempty"`
	IsActive                    bool     `json:"is_active"`
	CustomerID                  string   `json:"customer_id,omitempty"`
	Customer                    string   `json:"customer,omitempty"`
	TriggerID                   string   `json:"trigger_id,omitempty"`
	Trigger                     string   `json:"trigger,omitempty"`
	ServiceTypeUnitOfMeasureIDs []string `json:"service_type_unit_of_measure_ids,omitempty"`
}

func newServiceTypeUnitOfMeasureCohortsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service type unit of measure cohorts",
		Long: `List service type unit of measure cohorts with filtering and pagination.

Service type unit of measure cohorts group service type unit of measures
for a customer and define a trigger that selects the cohort.

Output Columns:
  ID           Cohort identifier
  NAME         Cohort name
  TRIGGER      Trigger service type unit of measure
  STUOM COUNT  Count of service type unit of measures in the cohort
  ACTIVE       Whether the cohort is active
  CUSTOMER     Customer name

Filters:
  --customer                        Filter by customer ID (comma-separated for multiple)
  --service-type-unit-of-measure-id Filter by service type unit of measure ID (matches cohorts containing the ID)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List cohorts
  xbe view service-type-unit-of-measure-cohorts list

  # Filter by customer
  xbe view service-type-unit-of-measure-cohorts list --customer 123

  # Filter by service type unit of measure
  xbe view service-type-unit-of-measure-cohorts list --service-type-unit-of-measure-id 456

  # Output as JSON
  xbe view service-type-unit-of-measure-cohorts list --json`,
		Args: cobra.NoArgs,
		RunE: runServiceTypeUnitOfMeasureCohortsList,
	}
	initServiceTypeUnitOfMeasureCohortsListFlags(cmd)
	return cmd
}

func init() {
	serviceTypeUnitOfMeasureCohortsCmd.AddCommand(newServiceTypeUnitOfMeasureCohortsListCmd())
}

func initServiceTypeUnitOfMeasureCohortsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("customer", "", "Filter by customer ID (comma-separated for multiple)")
	cmd.Flags().String("service-type-unit-of-measure-id", "", "Filter by service type unit of measure ID (comma-separated for multiple)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runServiceTypeUnitOfMeasureCohortsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseServiceTypeUnitOfMeasureCohortsListOptions(cmd)
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
	query.Set("fields[service-type-unit-of-measure-cohorts]", "name,is-active,service-type-unit-of-measure-ids,customer,trigger")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[service-type-unit-of-measures]", "name")
	query.Set("include", "customer,trigger")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[service-type-unit-of-measure-id]", opts.ServiceTypeUnitOfMeasureID)

	body, _, err := client.Get(cmd.Context(), "/v1/service-type-unit-of-measure-cohorts", query)
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

	rows := buildServiceTypeUnitOfMeasureCohortRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderServiceTypeUnitOfMeasureCohortsTable(cmd, rows)
}

func parseServiceTypeUnitOfMeasureCohortsListOptions(cmd *cobra.Command) (serviceTypeUnitOfMeasureCohortsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	customer, _ := cmd.Flags().GetString("customer")
	serviceTypeUnitOfMeasureID, _ := cmd.Flags().GetString("service-type-unit-of-measure-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return serviceTypeUnitOfMeasureCohortsListOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		NoAuth:                     noAuth,
		Limit:                      limit,
		Offset:                     offset,
		Sort:                       sort,
		Customer:                   customer,
		ServiceTypeUnitOfMeasureID: serviceTypeUnitOfMeasureID,
	}, nil
}

func buildServiceTypeUnitOfMeasureCohortRows(resp jsonAPIResponse) []serviceTypeUnitOfMeasureCohortRow {
	included := map[string]map[string]any{}
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc.Attributes
	}

	rows := make([]serviceTypeUnitOfMeasureCohortRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildServiceTypeUnitOfMeasureCohortRow(resource, included)
		rows = append(rows, row)
	}
	return rows
}

func buildServiceTypeUnitOfMeasureCohortRow(resource jsonAPIResource, included map[string]map[string]any) serviceTypeUnitOfMeasureCohortRow {
	row := serviceTypeUnitOfMeasureCohortRow{
		ID:                          resource.ID,
		Name:                        stringAttr(resource.Attributes, "name"),
		IsActive:                    boolAttr(resource.Attributes, "is-active"),
		ServiceTypeUnitOfMeasureIDs: stringSliceAttr(resource.Attributes, "service-type-unit-of-measure-ids"),
	}

	row.CustomerID = relationshipIDFromMap(resource.Relationships, "customer")
	row.TriggerID = relationshipIDFromMap(resource.Relationships, "trigger")
	row.Customer = resolveServiceTypeUnitOfMeasureCohortCustomerName(row.CustomerID, included)
	row.Trigger = resolveServiceTypeUnitOfMeasureCohortTriggerName(row.TriggerID, included)

	return row
}

func renderServiceTypeUnitOfMeasureCohortsTable(cmd *cobra.Command, rows []serviceTypeUnitOfMeasureCohortRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No service type unit of measure cohorts found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tTRIGGER\tSTUOM COUNT\tACTIVE\tCUSTOMER")
	for _, row := range rows {
		trigger := firstNonEmpty(row.Trigger, row.TriggerID)
		customer := firstNonEmpty(row.Customer, row.CustomerID)
		active := "no"
		if row.IsActive {
			active = "yes"
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%d\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			truncateString(trigger, 30),
			len(row.ServiceTypeUnitOfMeasureIDs),
			active,
			truncateString(customer, 30),
		)
	}

	return writer.Flush()
}

func resolveServiceTypeUnitOfMeasureCohortCustomerName(id string, included map[string]map[string]any) string {
	if id == "" {
		return ""
	}
	if attrs, ok := included[resourceKey("customers", id)]; ok {
		return firstNonEmpty(
			stringAttr(attrs, "company-name"),
			stringAttr(attrs, "name"),
		)
	}
	return ""
}

func resolveServiceTypeUnitOfMeasureCohortTriggerName(id string, included map[string]map[string]any) string {
	if id == "" {
		return ""
	}
	if attrs, ok := included[resourceKey("service-type-unit-of-measures", id)]; ok {
		return stringAttr(attrs, "name")
	}
	return ""
}
