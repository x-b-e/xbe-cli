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

type rootCausesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Incident     string
	IncidentType string
	IncidentID   string
	IsTriaged    string
}

type rootCauseRow struct {
	ID           string `json:"id"`
	Title        string `json:"title,omitempty"`
	IsTriaged    bool   `json:"is_triaged"`
	IncidentType string `json:"incident_type,omitempty"`
	IncidentID   string `json:"incident_id,omitempty"`
	RootCauseID  string `json:"root_cause_id,omitempty"`
}

func newRootCausesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List root causes",
		Long: `List root causes with filtering and pagination.

Output Columns:
  ID        Root cause identifier
  TITLE     Root cause title
  TRIAGED   Whether the root cause is triaged
  INCIDENT  Incident type and ID
  PARENT    Parent root cause ID

Filters:
  --incident         Filter by incident reference (Type|ID)
  --incident-type    Filter by incident type (use with --incident-id)
  --incident-id      Filter by incident ID (requires --incident-type)
  --is-triaged        Filter by triage status (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List root causes
  xbe view root-causes list

  # Filter by incident
  xbe view root-causes list --incident-type production-incidents --incident-id 123

  # Filter by triaged status
  xbe view root-causes list --is-triaged true

  # Output as JSON
  xbe view root-causes list --json`,
		Args: cobra.NoArgs,
		RunE: runRootCausesList,
	}
	initRootCausesListFlags(cmd)
	return cmd
}

func init() {
	rootCausesCmd.AddCommand(newRootCausesListCmd())
}

func initRootCausesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("incident", "", "Filter by incident reference (Type|ID)")
	cmd.Flags().String("incident-type", "", "Filter by incident type")
	cmd.Flags().String("incident-id", "", "Filter by incident ID (requires --incident-type)")
	cmd.Flags().String("is-triaged", "", "Filter by triage status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRootCausesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRootCausesListOptions(cmd)
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

	if opts.Incident != "" && (opts.IncidentType != "" || opts.IncidentID != "") {
		err := fmt.Errorf("--incident cannot be combined with --incident-type or --incident-id")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[root-causes]", "title,is-triaged,incident,root-cause")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	if opts.Incident != "" {
		incidentValue := strings.TrimSpace(opts.Incident)
		if parts := strings.SplitN(incidentValue, "|", 2); len(parts) == 2 {
			incidentType := normalizeIncidentTypeFilter(parts[0])
			incidentID := strings.TrimSpace(parts[1])
			incidentValue = incidentType + "|" + incidentID
		}
		query.Set("filter[incident]", incidentValue)
	} else {
		normalizedIncidentType := normalizeIncidentTypeFilter(opts.IncidentType)
		if opts.IncidentType != "" && opts.IncidentID != "" {
			query.Set("filter[incident]", normalizedIncidentType+"|"+opts.IncidentID)
		} else if opts.IncidentType != "" {
			query.Set("filter[incident_type]", normalizedIncidentType)
		} else if opts.IncidentID != "" {
			err := fmt.Errorf("--incident-id requires --incident-type")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}
	setFilterIfPresent(query, "filter[is-triaged]", opts.IsTriaged)

	body, _, err := client.Get(cmd.Context(), "/v1/root-causes", query)
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

	rows := buildRootCauseRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRootCausesTable(cmd, rows)
}

func parseRootCausesListOptions(cmd *cobra.Command) (rootCausesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	incident, _ := cmd.Flags().GetString("incident")
	incidentType, _ := cmd.Flags().GetString("incident-type")
	incidentID, _ := cmd.Flags().GetString("incident-id")
	isTriaged, _ := cmd.Flags().GetString("is-triaged")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rootCausesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Incident:     incident,
		IncidentType: incidentType,
		IncidentID:   incidentID,
		IsTriaged:    isTriaged,
	}, nil
}

func buildRootCauseRows(resp jsonAPIResponse) []rootCauseRow {
	rows := make([]rootCauseRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildRootCauseRow(resource))
	}
	return rows
}

func buildRootCauseRow(resource jsonAPIResource) rootCauseRow {
	row := rootCauseRow{
		ID:        resource.ID,
		Title:     stringAttr(resource.Attributes, "title"),
		IsTriaged: boolAttr(resource.Attributes, "is-triaged"),
	}

	if rel, ok := resource.Relationships["incident"]; ok && rel.Data != nil {
		row.IncidentType = rel.Data.Type
		row.IncidentID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["root-cause"]; ok && rel.Data != nil {
		row.RootCauseID = rel.Data.ID
	}

	return row
}

func renderRootCausesTable(cmd *cobra.Command, rows []rootCauseRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No root causes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTITLE\tTRIAGED\tINCIDENT\tPARENT")
	for _, row := range rows {
		triaged := "no"
		if row.IsTriaged {
			triaged = "yes"
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Title, 30),
			triaged,
			truncateString(formatIncidentReference(row.IncidentType, row.IncidentID), 30),
			row.RootCauseID,
		)
	}
	return writer.Flush()
}
