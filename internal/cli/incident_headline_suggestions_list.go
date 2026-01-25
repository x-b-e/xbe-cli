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

type incidentHeadlineSuggestionsListOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	NoAuth   bool
	Limit    int
	Offset   int
	Sort     string
	Incident string
}

type incidentHeadlineSuggestionRow struct {
	ID          string `json:"id"`
	IncidentID  string `json:"incident_id,omitempty"`
	IsAsync     bool   `json:"is_async"`
	IsFulfilled bool   `json:"is_fulfilled"`
	Suggestion  string `json:"suggestion,omitempty"`
}

func newIncidentHeadlineSuggestionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List incident headline suggestions",
		Long: `List incident headline suggestions with filtering and pagination.

Output Columns:
  ID         Suggestion identifier
  INCIDENT   Incident ID
  ASYNC      Whether suggestion is generated asynchronously
  FULFILLED  Whether the suggestion has been generated
  SUGGESTION Generated headline suggestion (truncated)

Filters:
  --incident  Filter by incident ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List incident headline suggestions
  xbe view incident-headline-suggestions list

  # Filter by incident
  xbe view incident-headline-suggestions list --incident 123

  # Output as JSON
  xbe view incident-headline-suggestions list --json`,
		Args: cobra.NoArgs,
		RunE: runIncidentHeadlineSuggestionsList,
	}
	initIncidentHeadlineSuggestionsListFlags(cmd)
	return cmd
}

func init() {
	incidentHeadlineSuggestionsCmd.AddCommand(newIncidentHeadlineSuggestionsListCmd())
}

func initIncidentHeadlineSuggestionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("incident", "", "Filter by incident ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIncidentHeadlineSuggestionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseIncidentHeadlineSuggestionsListOptions(cmd)
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
	query.Set("fields[incident-headline-suggestions]", "incident,is-async,is-fulfilled,suggestion")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[incident]", opts.Incident)

	body, _, err := client.Get(cmd.Context(), "/v1/incident-headline-suggestions", query)
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

	rows := buildIncidentHeadlineSuggestionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderIncidentHeadlineSuggestionsTable(cmd, rows)
}

func parseIncidentHeadlineSuggestionsListOptions(cmd *cobra.Command) (incidentHeadlineSuggestionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	incident, _ := cmd.Flags().GetString("incident")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return incidentHeadlineSuggestionsListOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		NoAuth:   noAuth,
		Limit:    limit,
		Offset:   offset,
		Sort:     sort,
		Incident: incident,
	}, nil
}

func buildIncidentHeadlineSuggestionRows(resp jsonAPIResponse) []incidentHeadlineSuggestionRow {
	rows := make([]incidentHeadlineSuggestionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildIncidentHeadlineSuggestionRow(resource))
	}
	return rows
}

func incidentHeadlineSuggestionRowFromSingle(resp jsonAPISingleResponse) incidentHeadlineSuggestionRow {
	return buildIncidentHeadlineSuggestionRow(resp.Data)
}

func buildIncidentHeadlineSuggestionRow(resource jsonAPIResource) incidentHeadlineSuggestionRow {
	row := incidentHeadlineSuggestionRow{
		ID:          resource.ID,
		IsAsync:     boolAttr(resource.Attributes, "is-async"),
		IsFulfilled: boolAttr(resource.Attributes, "is-fulfilled"),
		Suggestion:  stringAttr(resource.Attributes, "suggestion"),
	}

	if rel, ok := resource.Relationships["incident"]; ok && rel.Data != nil {
		row.IncidentID = rel.Data.ID
	}

	return row
}

func renderIncidentHeadlineSuggestionsTable(cmd *cobra.Command, rows []incidentHeadlineSuggestionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No incident headline suggestions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tINCIDENT\tASYNC\tFULFILLED\tSUGGESTION")
	for _, row := range rows {
		incidentID := row.IncidentID
		if incidentID == "" {
			incidentID = "-"
		}
		fmt.Fprintf(writer, "%s\t%s\t%t\t%t\t%s\n",
			row.ID,
			incidentID,
			row.IsAsync,
			row.IsFulfilled,
			truncateString(row.Suggestion, 40),
		)
	}
	return writer.Flush()
}
