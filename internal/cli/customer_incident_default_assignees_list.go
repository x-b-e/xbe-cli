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

type customerIncidentDefaultAssigneesListOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	NoAuth          bool
	Limit           int
	Offset          int
	Sort            string
	Customer        string
	DefaultAssignee string
	Kind            string
}

type customerIncidentDefaultAssigneeRow struct {
	ID                string `json:"id"`
	CustomerID        string `json:"customer_id,omitempty"`
	DefaultAssigneeID string `json:"default_assignee_id,omitempty"`
	Kind              string `json:"kind,omitempty"`
}

func newCustomerIncidentDefaultAssigneesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List customer incident default assignees",
		Long: `List customer incident default assignees.

Output Columns:
  ID                 Default assignee identifier
  CUSTOMER           Customer ID
  DEFAULT ASSIGNEE   User ID for the default assignee
  KIND               Incident kind

Filters:
  --customer          Filter by customer ID (comma-separated for multiple)
  --default-assignee  Filter by default assignee user ID (comma-separated for multiple)
  --kind              Filter by incident kind

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List default assignees
  xbe view customer-incident-default-assignees list

  # Filter by customer
  xbe view customer-incident-default-assignees list --customer 123

  # Filter by default assignee
  xbe view customer-incident-default-assignees list --default-assignee 456

  # Filter by kind
  xbe view customer-incident-default-assignees list --kind safety

  # Output as JSON
  xbe view customer-incident-default-assignees list --json`,
		Args: cobra.NoArgs,
		RunE: runCustomerIncidentDefaultAssigneesList,
	}
	initCustomerIncidentDefaultAssigneesListFlags(cmd)
	return cmd
}

func init() {
	customerIncidentDefaultAssigneesCmd.AddCommand(newCustomerIncidentDefaultAssigneesListCmd())
}

func initCustomerIncidentDefaultAssigneesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("customer", "", "Filter by customer ID (comma-separated for multiple)")
	cmd.Flags().String("default-assignee", "", "Filter by default assignee user ID (comma-separated for multiple)")
	cmd.Flags().String("kind", "", "Filter by incident kind")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomerIncidentDefaultAssigneesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCustomerIncidentDefaultAssigneesListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[customer-incident-default-assignees]", "kind,customer,default-assignee")

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
	setFilterIfPresent(query, "filter[default-assignee]", opts.DefaultAssignee)
	setFilterIfPresent(query, "filter[kind]", opts.Kind)

	body, _, err := client.Get(cmd.Context(), "/v1/customer-incident-default-assignees", query)
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

	rows := buildCustomerIncidentDefaultAssigneeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCustomerIncidentDefaultAssigneesTable(cmd, rows)
}

func parseCustomerIncidentDefaultAssigneesListOptions(cmd *cobra.Command) (customerIncidentDefaultAssigneesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	customer, _ := cmd.Flags().GetString("customer")
	defaultAssignee, _ := cmd.Flags().GetString("default-assignee")
	kind, _ := cmd.Flags().GetString("kind")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return customerIncidentDefaultAssigneesListOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		NoAuth:          noAuth,
		Limit:           limit,
		Offset:          offset,
		Sort:            sort,
		Customer:        customer,
		DefaultAssignee: defaultAssignee,
		Kind:            kind,
	}, nil
}

func buildCustomerIncidentDefaultAssigneeRows(resp jsonAPIResponse) []customerIncidentDefaultAssigneeRow {
	rows := make([]customerIncidentDefaultAssigneeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, customerIncidentDefaultAssigneeRowFromResource(resource))
	}
	return rows
}

func customerIncidentDefaultAssigneeRowFromResource(resource jsonAPIResource) customerIncidentDefaultAssigneeRow {
	return customerIncidentDefaultAssigneeRow{
		ID:                resource.ID,
		CustomerID:        relationshipIDFromMap(resource.Relationships, "customer"),
		DefaultAssigneeID: relationshipIDFromMap(resource.Relationships, "default-assignee"),
		Kind:              stringAttr(resource.Attributes, "kind"),
	}
}

func renderCustomerIncidentDefaultAssigneesTable(cmd *cobra.Command, rows []customerIncidentDefaultAssigneeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No customer incident default assignees found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCUSTOMER\tDEFAULT ASSIGNEE\tKIND")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n", row.ID, row.CustomerID, row.DefaultAssigneeID, row.Kind)
	}
	return writer.Flush()
}
