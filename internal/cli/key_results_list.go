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

type keyResultsListOptions struct {
	BaseURL                             string
	Token                               string
	JSON                                bool
	NoAuth                              bool
	Limit                               int
	Offset                              int
	Sort                                string
	Owner                               string
	Objective                           string
	Status                              string
	IsTemplate                          string
	CustomerSuccessResponsiblePerson    string
	HasCustomerSuccessResponsiblePerson string
}

type keyResultRow struct {
	ID                                   string `json:"id"`
	Title                                string `json:"title,omitempty"`
	Status                               string `json:"status,omitempty"`
	StartOn                              string `json:"start_on,omitempty"`
	EndOn                                string `json:"end_on,omitempty"`
	CompletionPercentage                 any    `json:"completion_percentage,omitempty"`
	ObjectiveID                          string `json:"objective_id,omitempty"`
	ObjectiveName                        string `json:"objective_name,omitempty"`
	OwnerID                              string `json:"owner_id,omitempty"`
	OwnerName                            string `json:"owner_name,omitempty"`
	CustomerSuccessResponsiblePersonID   string `json:"customer_success_responsible_person_id,omitempty"`
	CustomerSuccessResponsiblePersonName string `json:"customer_success_responsible_person_name,omitempty"`
}

func newKeyResultsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List key results",
		Long: `List key results with filtering and pagination.

Output Columns:
  ID           Key result identifier
  STATUS       Current status
  TITLE        Key result title
  COMPLETION   Completion percentage (0-1)
  OBJECTIVE    Objective name (or ID)
  OWNER        Owner name (or ID)

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filters:
  --owner                              Filter by owner user ID
  --objective                          Filter by objective ID
  --status                             Filter by status (unknown, not_started, red, yellow, green, completed, scrapped)
  --is-template                        Filter by objective template status (true/false)
  --customer-success-responsible-person Filter by customer success responsible person user ID
  --has-customer-success-responsible-person Filter by whether a customer success responsible person is set (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List key results
  xbe view key-results list

  # Filter by objective
  xbe view key-results list --objective 123

  # Filter by owner
  xbe view key-results list --owner 456

  # Filter by status
  xbe view key-results list --status green

  # Filter by objective template status
  xbe view key-results list --is-template true

  # Output as JSON
  xbe view key-results list --json`,
		Args: cobra.NoArgs,
		RunE: runKeyResultsList,
	}
	initKeyResultsListFlags(cmd)
	return cmd
}

func init() {
	keyResultsCmd.AddCommand(newKeyResultsListCmd())
}

func initKeyResultsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("owner", "", "Filter by owner user ID (comma-separated for multiple)")
	cmd.Flags().String("objective", "", "Filter by objective ID (comma-separated for multiple)")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("is-template", "", "Filter by objective template status (true/false)")
	cmd.Flags().String("customer-success-responsible-person", "", "Filter by customer success responsible person user ID (comma-separated for multiple)")
	cmd.Flags().String("has-customer-success-responsible-person", "", "Filter by whether customer success responsible person is set (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runKeyResultsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseKeyResultsListOptions(cmd)
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
	query.Set("fields[key-results]", "title,status,start-on,end-on,completion-percentage,objective,owner,customer-success-responsible-person")
	query.Set("include", "objective,owner,customer-success-responsible-person")
	query.Set("fields[objectives]", "name")
	query.Set("fields[users]", "name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[owner]", opts.Owner)
	setFilterIfPresent(query, "filter[objective]", opts.Objective)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[is-template]", opts.IsTemplate)
	setFilterIfPresent(query, "filter[customer-success-responsible-person]", opts.CustomerSuccessResponsiblePerson)
	setFilterIfPresent(query, "filter[has-customer-success-responsible-person]", opts.HasCustomerSuccessResponsiblePerson)

	body, _, err := client.Get(cmd.Context(), "/v1/key-results", query)
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

	rows := buildKeyResultRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderKeyResultsTable(cmd, rows)
}

func parseKeyResultsListOptions(cmd *cobra.Command) (keyResultsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	owner, _ := cmd.Flags().GetString("owner")
	objective, _ := cmd.Flags().GetString("objective")
	status, _ := cmd.Flags().GetString("status")
	isTemplate, _ := cmd.Flags().GetString("is-template")
	customerSuccessResponsiblePerson, _ := cmd.Flags().GetString("customer-success-responsible-person")
	hasCustomerSuccessResponsiblePerson, _ := cmd.Flags().GetString("has-customer-success-responsible-person")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return keyResultsListOptions{
		BaseURL:                             baseURL,
		Token:                               token,
		JSON:                                jsonOut,
		NoAuth:                              noAuth,
		Limit:                               limit,
		Offset:                              offset,
		Sort:                                sort,
		Owner:                               owner,
		Objective:                           objective,
		Status:                              status,
		IsTemplate:                          isTemplate,
		CustomerSuccessResponsiblePerson:    customerSuccessResponsiblePerson,
		HasCustomerSuccessResponsiblePerson: hasCustomerSuccessResponsiblePerson,
	}, nil
}

func buildKeyResultRows(resp jsonAPIResponse) []keyResultRow {
	included := indexIncludedResources(resp.Included)
	rows := make([]keyResultRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := keyResultRow{
			ID:                   resource.ID,
			Title:                strings.TrimSpace(stringAttr(resource.Attributes, "title")),
			Status:               stringAttr(resource.Attributes, "status"),
			StartOn:              formatDate(stringAttr(resource.Attributes, "start-on")),
			EndOn:                formatDate(stringAttr(resource.Attributes, "end-on")),
			CompletionPercentage: resource.Attributes["completion-percentage"],
		}

		if rel, ok := resource.Relationships["objective"]; ok && rel.Data != nil {
			row.ObjectiveID = rel.Data.ID
			if obj, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ObjectiveName = strings.TrimSpace(stringAttr(obj.Attributes, "name"))
			}
		}
		if rel, ok := resource.Relationships["owner"]; ok && rel.Data != nil {
			row.OwnerID = rel.Data.ID
			if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.OwnerName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			}
		}
		if rel, ok := resource.Relationships["customer-success-responsible-person"]; ok && rel.Data != nil {
			row.CustomerSuccessResponsiblePersonID = rel.Data.ID
			if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.CustomerSuccessResponsiblePersonName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderKeyResultsTable(cmd *cobra.Command, rows []keyResultRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No key results found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tTITLE\tCOMPLETION\tOBJECTIVE\tOWNER")

	for _, row := range rows {
		objective := firstNonEmpty(row.ObjectiveName, row.ObjectiveID)
		owner := firstNonEmpty(row.OwnerName, row.OwnerID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			truncateString(row.Title, 40),
			formatAnyValue(row.CompletionPercentage),
			truncateString(objective, 30),
			truncateString(owner, 25),
		)
	}

	return writer.Flush()
}
