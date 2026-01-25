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

type commitmentSimulationSetsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	Organization        string
	OrganizationID      string
	OrganizationType    string
	NotOrganizationType string
	Status              string
}

type commitmentSimulationSetRow struct {
	ID               string `json:"id"`
	Status           string `json:"status,omitempty"`
	StartOn          string `json:"start_on,omitempty"`
	EndOn            string `json:"end_on,omitempty"`
	IterationCount   string `json:"iteration_count,omitempty"`
	OrganizationType string `json:"organization_type,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
}

func newCommitmentSimulationSetsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List commitment simulation sets",
		Long: `List commitment simulation sets.

Output Columns:
  ID              Commitment simulation set identifier
  STATUS          Processing status (enqueued or processed)
  START           Start date
  END             End date
  ITERATIONS      Iteration count
  ORGANIZATION    Organization type and ID

Filters:
  --organization           Filter by organization (format: Type|ID, e.g., Broker|123)
  --organization-id        Filter by organization ID (requires --organization-type or Type|ID)
  --organization-type      Filter by organization type (e.g., Broker, Customer, Trucker)
  --not-organization-type  Exclude organization type (e.g., Customer)
  --status                 Filter by status (enqueued or processed)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List commitment simulation sets
  xbe view commitment-simulation-sets list

  # Filter by organization
  xbe view commitment-simulation-sets list --organization "Broker|123"

  # Filter by organization type + id
  xbe view commitment-simulation-sets list --organization-type Broker --organization-id 123

  # Filter by status
  xbe view commitment-simulation-sets list --status enqueued

  # Output as JSON
  xbe view commitment-simulation-sets list --json`,
		Args: cobra.NoArgs,
		RunE: runCommitmentSimulationSetsList,
	}
	initCommitmentSimulationSetsListFlags(cmd)
	return cmd
}

func init() {
	commitmentSimulationSetsCmd.AddCommand(newCommitmentSimulationSetsListCmd())
}

func initCommitmentSimulationSetsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("organization", "", "Filter by organization (format: Type|ID, e.g., Broker|123)")
	cmd.Flags().String("organization-id", "", "Filter by organization ID (requires --organization-type or Type|ID)")
	cmd.Flags().String("organization-type", "", "Filter by organization type (e.g., Broker, Customer, Trucker)")
	cmd.Flags().String("not-organization-type", "", "Exclude organization type (e.g., Customer)")
	cmd.Flags().String("status", "", "Filter by status (enqueued or processed)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCommitmentSimulationSetsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCommitmentSimulationSetsListOptions(cmd)
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
	query.Set("fields[commitment-simulation-sets]", "start-on,end-on,iteration-count,status,organization")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[organization]", opts.Organization)
	organizationIDFilter, err := buildOrganizationIDFilter(opts.OrganizationType, opts.OrganizationID)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if organizationIDFilter != "" {
		query.Set("filter[organization_id]", organizationIDFilter)
	} else {
		setFilterIfPresent(query, "filter[organization_type]", opts.OrganizationType)
	}
	setFilterIfPresent(query, "filter[status]", opts.Status)

	body, _, err := client.Get(cmd.Context(), "/v1/commitment-simulation-sets", query)
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

	rows := buildCommitmentSimulationSetRows(resp)
	if strings.TrimSpace(opts.NotOrganizationType) != "" {
		rows = filterCommitmentSimulationSetsByOrganizationType(rows, opts.NotOrganizationType)
	}
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCommitmentSimulationSetsTable(cmd, rows)
}

func parseCommitmentSimulationSetsListOptions(cmd *cobra.Command) (commitmentSimulationSetsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	organization, _ := cmd.Flags().GetString("organization")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	notOrganizationType, _ := cmd.Flags().GetString("not-organization-type")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return commitmentSimulationSetsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		Organization:        organization,
		OrganizationID:      organizationID,
		OrganizationType:    organizationType,
		NotOrganizationType: notOrganizationType,
		Status:              status,
	}, nil
}

func buildCommitmentSimulationSetRows(resp jsonAPIResponse) []commitmentSimulationSetRow {
	rows := make([]commitmentSimulationSetRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := commitmentSimulationSetRow{
			ID:             resource.ID,
			Status:         stringAttr(resource.Attributes, "status"),
			StartOn:        formatDate(stringAttr(resource.Attributes, "start-on")),
			EndOn:          formatDate(stringAttr(resource.Attributes, "end-on")),
			IterationCount: stringAttr(resource.Attributes, "iteration-count"),
		}

		if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			row.OrganizationType = rel.Data.Type
			row.OrganizationID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderCommitmentSimulationSetsTable(cmd *cobra.Command, rows []commitmentSimulationSetRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No commitment simulation sets found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tSTART\tEND\tITERATIONS\tORGANIZATION")
	for _, row := range rows {
		organization := formatPolymorphic(row.OrganizationType, row.OrganizationID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.StartOn,
			row.EndOn,
			row.IterationCount,
			organization,
		)
	}
	return writer.Flush()
}

func buildOrganizationIDFilter(organizationType, organizationID string) (string, error) {
	organizationID = strings.TrimSpace(organizationID)
	if organizationID == "" {
		return "", nil
	}
	if strings.Contains(organizationID, "|") {
		return organizationID, nil
	}
	organizationType = strings.TrimSpace(organizationType)
	if organizationType == "" {
		return "", fmt.Errorf("--organization-type is required when using --organization-id (or use --organization Type|ID)")
	}
	return organizationType + "|" + organizationID, nil
}

func filterCommitmentSimulationSetsByOrganizationType(rows []commitmentSimulationSetRow, organizationType string) []commitmentSimulationSetRow {
	filterType := normalizeOrganizationType(organizationType)
	if filterType == "" {
		return rows
	}
	filtered := make([]commitmentSimulationSetRow, 0, len(rows))
	for _, row := range rows {
		if normalizeOrganizationType(row.OrganizationType) != filterType {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func normalizeOrganizationType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "broker", "brokers":
		return "broker"
	case "customer", "customers":
		return "customer"
	case "trucker", "truckers":
		return "trucker"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}
