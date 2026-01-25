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

type projectTransportPlanEventLocationPredictionAutopsiesListOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	NoAuth                    bool
	Limit                     int
	Offset                    int
	Sort                      string
	ProjectTransportPlanEvent string
}

type projectTransportPlanEventLocationPredictionAutopsyRow struct {
	ID                          string `json:"id"`
	Status                      string `json:"status,omitempty"`
	Error                       string `json:"error,omitempty"`
	ProjectTransportPlanEventID string `json:"project_transport_plan_event_id,omitempty"`
}

func newProjectTransportPlanEventLocationPredictionAutopsiesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan event location prediction autopsies",
		Long: `List project transport plan event location prediction autopsies.

Output Columns:
  ID       Autopsy identifier
  STATUS   Autopsy status
  ERROR    Error summary (if present)
  EVENT    Project transport plan event ID

Filters:
  --project-transport-plan-event  Filter by project transport plan event ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List autopsies
  xbe view project-transport-plan-event-location-prediction-autopsies list --limit 10

  # Filter by project transport plan event
  xbe view project-transport-plan-event-location-prediction-autopsies list --project-transport-plan-event 123

  # Output as JSON
  xbe view project-transport-plan-event-location-prediction-autopsies list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanEventLocationPredictionAutopsiesList,
	}
	initProjectTransportPlanEventLocationPredictionAutopsiesListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanEventLocationPredictionAutopsiesCmd.AddCommand(newProjectTransportPlanEventLocationPredictionAutopsiesListCmd())
}

func initProjectTransportPlanEventLocationPredictionAutopsiesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-transport-plan-event", "", "Filter by project transport plan event ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanEventLocationPredictionAutopsiesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanEventLocationPredictionAutopsiesListOptions(cmd)
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
	query.Set("fields[project-transport-plan-event-location-prediction-autopsies]", "status,error,project-transport-plan-event")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[project-transport-plan-event]", opts.ProjectTransportPlanEvent)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-event-location-prediction-autopsies", query)
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

	rows := buildProjectTransportPlanEventLocationPredictionAutopsyRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanEventLocationPredictionAutopsiesTable(cmd, rows)
}

func parseProjectTransportPlanEventLocationPredictionAutopsiesListOptions(cmd *cobra.Command) (projectTransportPlanEventLocationPredictionAutopsiesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectTransportPlanEvent, _ := cmd.Flags().GetString("project-transport-plan-event")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanEventLocationPredictionAutopsiesListOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		NoAuth:                    noAuth,
		Limit:                     limit,
		Offset:                    offset,
		Sort:                      sort,
		ProjectTransportPlanEvent: projectTransportPlanEvent,
	}, nil
}

func buildProjectTransportPlanEventLocationPredictionAutopsyRows(resp jsonAPIResponse) []projectTransportPlanEventLocationPredictionAutopsyRow {
	rows := make([]projectTransportPlanEventLocationPredictionAutopsyRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildProjectTransportPlanEventLocationPredictionAutopsyRow(resource))
	}
	return rows
}

func buildProjectTransportPlanEventLocationPredictionAutopsyRow(resource jsonAPIResource) projectTransportPlanEventLocationPredictionAutopsyRow {
	attrs := resource.Attributes
	row := projectTransportPlanEventLocationPredictionAutopsyRow{
		ID:     resource.ID,
		Status: stringAttr(attrs, "status"),
		Error:  stringAttr(attrs, "error"),
	}

	if rel, ok := resource.Relationships["project-transport-plan-event"]; ok && rel.Data != nil {
		row.ProjectTransportPlanEventID = rel.Data.ID
	}

	return row
}

func renderProjectTransportPlanEventLocationPredictionAutopsiesTable(cmd *cobra.Command, rows []projectTransportPlanEventLocationPredictionAutopsyRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan event location prediction autopsies found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tERROR\tEVENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			truncateString(row.Error, 60),
			row.ProjectTransportPlanEventID,
		)
	}
	return writer.Flush()
}
