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

type projectCustomersListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Project      string
	Customer     string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
}

type projectCustomerRow struct {
	ID        string `json:"id"`
	Project   string `json:"project_id,omitempty"`
	Customer  string `json:"customer_id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func newProjectCustomersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project customers",
		Long: `List project customers with filtering and pagination.

Output Columns:
  ID         Project customer identifier
  PROJECT    Project ID
  CUSTOMER   Customer ID
  CREATED AT Creation timestamp
  UPDATED AT Last update timestamp

Filters:
  --project         Filter by project ID
  --customer        Filter by customer ID
  --created-at-min  Filter by created-at on/after (ISO 8601)
  --created-at-max  Filter by created-at on/before (ISO 8601)
  --updated-at-min  Filter by updated-at on/after (ISO 8601)
  --updated-at-max  Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project customers
  xbe view project-customers list

  # Filter by project
  xbe view project-customers list --project 123

  # Filter by customer
  xbe view project-customers list --customer 456

  # Output as JSON
  xbe view project-customers list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectCustomersList,
	}
	initProjectCustomersListFlags(cmd)
	return cmd
}

func init() {
	projectCustomersCmd.AddCommand(newProjectCustomersListCmd())
}

func initProjectCustomersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectCustomersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectCustomersListOptions(cmd)
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
	query.Set("fields[project-customers]", "created-at,updated-at,project,customer")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/project-customers", query)
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

	rows := buildProjectCustomerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectCustomersTable(cmd, rows)
}

func parseProjectCustomersListOptions(cmd *cobra.Command) (projectCustomersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	project, _ := cmd.Flags().GetString("project")
	customer, _ := cmd.Flags().GetString("customer")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectCustomersListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Project:      project,
		Customer:     customer,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
	}, nil
}

func buildProjectCustomerRows(resp jsonAPIResponse) []projectCustomerRow {
	rows := make([]projectCustomerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := projectCustomerRow{
			ID:        resource.ID,
			CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
			UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
		}
		if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
			row.Project = rel.Data.ID
		}
		if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
			row.Customer = rel.Data.ID
		}
		rows = append(rows, row)
	}
	return rows
}

func renderProjectCustomersTable(cmd *cobra.Command, rows []projectCustomerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project customers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPROJECT\tCUSTOMER\tCREATED AT\tUPDATED AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Project,
			row.Customer,
			row.CreatedAt,
			row.UpdatedAt,
		)
	}
	return writer.Flush()
}
