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

type projectTransportPlanStrategyStepsListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	Sort      string
	Position  string
	Strategy  string
	EventType string
}

type projectTransportPlanStrategyStepRow struct {
	ID        string `json:"id"`
	Position  int    `json:"position"`
	Strategy  string `json:"strategy_id,omitempty"`
	EventType string `json:"event_type_id,omitempty"`
}

func newProjectTransportPlanStrategyStepsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan strategy steps",
		Long: `List project transport plan strategy steps with filtering and pagination.

Output Columns:
  ID         Project transport plan strategy step identifier
  POSITION   Step position within the strategy
  STRATEGY   Project transport plan strategy ID
  EVENT TYPE Project transport event type ID

Filters:
  --position   Filter by step position
  --strategy   Filter by project transport plan strategy ID
  --event-type Filter by project transport event type ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project transport plan strategy steps
  xbe view project-transport-plan-strategy-steps list

  # Filter by strategy
  xbe view project-transport-plan-strategy-steps list --strategy 123

  # Filter by event type
  xbe view project-transport-plan-strategy-steps list --event-type 456

  # Output as JSON
  xbe view project-transport-plan-strategy-steps list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanStrategyStepsList,
	}
	initProjectTransportPlanStrategyStepsListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanStrategyStepsCmd.AddCommand(newProjectTransportPlanStrategyStepsListCmd())
}

func initProjectTransportPlanStrategyStepsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("position", "", "Filter by step position")
	cmd.Flags().String("strategy", "", "Filter by project transport plan strategy ID")
	cmd.Flags().String("event-type", "", "Filter by project transport event type ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanStrategyStepsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanStrategyStepsListOptions(cmd)
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
	query.Set("fields[project-transport-plan-strategy-steps]", "position,strategy,event-type")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[position]", opts.Position)
	setFilterIfPresent(query, "filter[strategy]", opts.Strategy)
	setFilterIfPresent(query, "filter[event-type]", opts.EventType)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-strategy-steps", query)
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

	rows := buildProjectTransportPlanStrategyStepRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanStrategyStepsTable(cmd, rows)
}

func parseProjectTransportPlanStrategyStepsListOptions(cmd *cobra.Command) (projectTransportPlanStrategyStepsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	position, _ := cmd.Flags().GetString("position")
	strategy, _ := cmd.Flags().GetString("strategy")
	eventType, _ := cmd.Flags().GetString("event-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanStrategyStepsListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		Sort:      sort,
		Position:  position,
		Strategy:  strategy,
		EventType: eventType,
	}, nil
}

func buildProjectTransportPlanStrategyStepRows(resp jsonAPIResponse) []projectTransportPlanStrategyStepRow {
	rows := make([]projectTransportPlanStrategyStepRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := projectTransportPlanStrategyStepRow{
			ID:       resource.ID,
			Position: intAttr(attrs, "position"),
		}
		if rel, ok := resource.Relationships["strategy"]; ok && rel.Data != nil {
			row.Strategy = rel.Data.ID
		}
		if rel, ok := resource.Relationships["event-type"]; ok && rel.Data != nil {
			row.EventType = rel.Data.ID
		}
		rows = append(rows, row)
	}
	return rows
}

func renderProjectTransportPlanStrategyStepsTable(cmd *cobra.Command, rows []projectTransportPlanStrategyStepRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan strategy steps found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPOSITION\tSTRATEGY\tEVENT TYPE")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%d\t%s\t%s\n",
			row.ID,
			row.Position,
			row.Strategy,
			row.EventType,
		)
	}
	return writer.Flush()
}
