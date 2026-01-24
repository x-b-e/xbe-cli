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

type projectTransportPlanEventLocationPredictionsListOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	NoAuth                    bool
	Limit                     int
	Offset                    int
	Sort                      string
	ProjectTransportPlanEvent string
	TransportOrder            string
}

type projectTransportPlanEventLocationPredictionRow struct {
	ID                          string `json:"id"`
	ProjectTransportPlanEventID string `json:"project_transport_plan_event_id,omitempty"`
	TransportOrderID            string `json:"transport_order_id,omitempty"`
	PredictionCount             int    `json:"prediction_count,omitempty"`
}

func newProjectTransportPlanEventLocationPredictionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan event location predictions",
		Long: `List project transport plan event location predictions.

Output Columns:
  ID           Prediction identifier
  EVENT        Project transport plan event ID
  ORDER        Transport order ID
  PREDICTIONS  Prediction count

Filters:
  --project-transport-plan-event  Filter by project transport plan event ID
  --transport-order               Filter by transport order ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List predictions
  xbe view project-transport-plan-event-location-predictions list

  # Filter by event
  xbe view project-transport-plan-event-location-predictions list --project-transport-plan-event 123

  # JSON output
  xbe view project-transport-plan-event-location-predictions list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanEventLocationPredictionsList,
	}
	initProjectTransportPlanEventLocationPredictionsListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanEventLocationPredictionsCmd.AddCommand(newProjectTransportPlanEventLocationPredictionsListCmd())
}

func initProjectTransportPlanEventLocationPredictionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-transport-plan-event", "", "Filter by project transport plan event ID")
	cmd.Flags().String("transport-order", "", "Filter by transport order ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanEventLocationPredictionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanEventLocationPredictionsListOptions(cmd)
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
	query.Set("fields[project-transport-plan-event-location-predictions]", "predictions,project-transport-plan-event,transport-order")

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
	setFilterIfPresent(query, "filter[transport-order]", opts.TransportOrder)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-event-location-predictions", query)
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

	rows := buildProjectTransportPlanEventLocationPredictionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanEventLocationPredictionsTable(cmd, rows)
}

func parseProjectTransportPlanEventLocationPredictionsListOptions(cmd *cobra.Command) (projectTransportPlanEventLocationPredictionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectTransportPlanEvent, _ := cmd.Flags().GetString("project-transport-plan-event")
	transportOrder, _ := cmd.Flags().GetString("transport-order")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanEventLocationPredictionsListOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		NoAuth:                    noAuth,
		Limit:                     limit,
		Offset:                    offset,
		Sort:                      sort,
		ProjectTransportPlanEvent: projectTransportPlanEvent,
		TransportOrder:            transportOrder,
	}, nil
}

func buildProjectTransportPlanEventLocationPredictionRows(resp jsonAPIResponse) []projectTransportPlanEventLocationPredictionRow {
	rows := make([]projectTransportPlanEventLocationPredictionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectTransportPlanEventLocationPredictionRow{
			ID:              resource.ID,
			PredictionCount: countLocationPredictions(resource.Attributes),
		}

		if rel, ok := resource.Relationships["project-transport-plan-event"]; ok && rel.Data != nil {
			row.ProjectTransportPlanEventID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["transport-order"]; ok && rel.Data != nil {
			row.TransportOrderID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func countLocationPredictions(attrs map[string]any) int {
	if attrs == nil {
		return 0
	}
	value, ok := attrs["predictions"]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case []any:
		return len(typed)
	case []map[string]any:
		return len(typed)
	default:
		return 0
	}
}

func renderProjectTransportPlanEventLocationPredictionsTable(cmd *cobra.Command, rows []projectTransportPlanEventLocationPredictionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan event location predictions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tEVENT\tORDER\tPREDICTIONS")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%d\n",
			row.ID,
			row.ProjectTransportPlanEventID,
			row.TransportOrderID,
			row.PredictionCount,
		)
	}
	return writer.Flush()
}
