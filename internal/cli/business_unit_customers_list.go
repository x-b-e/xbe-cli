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

type businessUnitCustomersListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	BusinessUnit string
	Customer     string
}

type businessUnitCustomerRow struct {
	ID               string `json:"id"`
	BusinessUnitID   string `json:"business_unit_id,omitempty"`
	BusinessUnitName string `json:"business_unit_name,omitempty"`
	CustomerID       string `json:"customer_id,omitempty"`
	CustomerName     string `json:"customer_name,omitempty"`
}

func newBusinessUnitCustomersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List business unit customers",
		Long: `List business unit customers with filtering and pagination.

Output Columns:
  ID             Link identifier
  BUSINESS UNIT  Business unit name (falls back to ID)
  CUSTOMER       Customer name (falls back to ID)

Filters:
  --business-unit  Filter by business unit ID
  --customer       Filter by customer ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List business unit customers
  xbe view business-unit-customers list

  # Filter by business unit
  xbe view business-unit-customers list --business-unit 123

  # Filter by customer
  xbe view business-unit-customers list --customer 456

  # JSON output
  xbe view business-unit-customers list --json`,
		Args: cobra.NoArgs,
		RunE: runBusinessUnitCustomersList,
	}
	initBusinessUnitCustomersListFlags(cmd)
	return cmd
}

func init() {
	businessUnitCustomersCmd.AddCommand(newBusinessUnitCustomersListCmd())
}

func initBusinessUnitCustomersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("business-unit", "", "Filter by business unit ID")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBusinessUnitCustomersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseBusinessUnitCustomersListOptions(cmd)
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
	query.Set("fields[business-unit-customers]", "business-unit,customer")
	query.Set("fields[business-units]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("include", "business-unit,customer")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[business-unit]", opts.BusinessUnit)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)

	body, _, err := client.Get(cmd.Context(), "/v1/business-unit-customers", query)
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

	rows := buildBusinessUnitCustomerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderBusinessUnitCustomersTable(cmd, rows)
}

func parseBusinessUnitCustomersListOptions(cmd *cobra.Command) (businessUnitCustomersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	customer, _ := cmd.Flags().GetString("customer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return businessUnitCustomersListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		BusinessUnit: businessUnit,
		Customer:     customer,
	}, nil
}

func buildBusinessUnitCustomerRows(resp jsonAPIResponse) []businessUnitCustomerRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]businessUnitCustomerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := businessUnitCustomerRow{
			ID: resource.ID,
		}

		if rel, ok := resource.Relationships["business-unit"]; ok && rel.Data != nil {
			row.BusinessUnitID = rel.Data.ID
			if unit, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.BusinessUnitName = strings.TrimSpace(stringAttr(unit.Attributes, "company-name"))
			}
		}

		if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
			row.CustomerID = rel.Data.ID
			if customer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.CustomerName = strings.TrimSpace(stringAttr(customer.Attributes, "company-name"))
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderBusinessUnitCustomersTable(cmd *cobra.Command, rows []businessUnitCustomerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No business unit customers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tBUSINESS UNIT\tCUSTOMER")
	for _, row := range rows {
		businessUnitDisplay := firstNonEmpty(row.BusinessUnitName, row.BusinessUnitID)
		customerDisplay := firstNonEmpty(row.CustomerName, row.CustomerID)
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			truncateString(businessUnitDisplay, 40),
			truncateString(customerDisplay, 40),
		)
	}
	return writer.Flush()
}
