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

type projectTransportPlanSegmentDriversListOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	NoAuth                      bool
	Limit                       int
	Offset                      int
	Sort                        string
	ProjectTransportPlanSegment string
	Driver                      string
	CreatedAtMin                string
	CreatedAtMax                string
	UpdatedAtMin                string
	UpdatedAtMax                string
}

type projectTransportPlanSegmentDriverRow struct {
	ID                          string `json:"id"`
	ProjectTransportPlanSegment string `json:"project_transport_plan_segment_id,omitempty"`
	Driver                      string `json:"driver_id,omitempty"`
	CreatedAt                   string `json:"created_at,omitempty"`
	UpdatedAt                   string `json:"updated_at,omitempty"`
}

func newProjectTransportPlanSegmentDriversListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan segment drivers",
		Long: `List project transport plan segment drivers with filtering and pagination.

Output Columns:
  ID         Project transport plan segment driver identifier
  SEGMENT    Project transport plan segment ID
  DRIVER     Driver (user) ID
  CREATED AT Creation timestamp
  UPDATED AT Last update timestamp

Filters:
  --project-transport-plan-segment  Filter by project transport plan segment ID
  --driver                          Filter by driver (user) ID
  --created-at-min                  Filter by created-at on/after (ISO 8601)
  --created-at-max                  Filter by created-at on/before (ISO 8601)
  --updated-at-min                  Filter by updated-at on/after (ISO 8601)
  --updated-at-max                  Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project transport plan segment drivers
  xbe view project-transport-plan-segment-drivers list

  # Filter by segment
  xbe view project-transport-plan-segment-drivers list --project-transport-plan-segment 123

  # Filter by driver
  xbe view project-transport-plan-segment-drivers list --driver 456

  # Output as JSON
  xbe view project-transport-plan-segment-drivers list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanSegmentDriversList,
	}
	initProjectTransportPlanSegmentDriversListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanSegmentDriversCmd.AddCommand(newProjectTransportPlanSegmentDriversListCmd())
}

func initProjectTransportPlanSegmentDriversListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-transport-plan-segment", "", "Filter by project transport plan segment ID")
	cmd.Flags().String("driver", "", "Filter by driver (user) ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanSegmentDriversList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanSegmentDriversListOptions(cmd)
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
	query.Set("fields[project-transport-plan-segment-drivers]", "created-at,updated-at,project-transport-plan-segment,driver")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[project-transport-plan-segment]", opts.ProjectTransportPlanSegment)
	setFilterIfPresent(query, "filter[driver]", opts.Driver)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-segment-drivers", query)
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

	rows := buildProjectTransportPlanSegmentDriverRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanSegmentDriversTable(cmd, rows)
}

func parseProjectTransportPlanSegmentDriversListOptions(cmd *cobra.Command) (projectTransportPlanSegmentDriversListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	segment, _ := cmd.Flags().GetString("project-transport-plan-segment")
	driver, _ := cmd.Flags().GetString("driver")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanSegmentDriversListOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		NoAuth:                      noAuth,
		Limit:                       limit,
		Offset:                      offset,
		Sort:                        sort,
		ProjectTransportPlanSegment: segment,
		Driver:                      driver,
		CreatedAtMin:                createdAtMin,
		CreatedAtMax:                createdAtMax,
		UpdatedAtMin:                updatedAtMin,
		UpdatedAtMax:                updatedAtMax,
	}, nil
}

func buildProjectTransportPlanSegmentDriverRows(resp jsonAPIResponse) []projectTransportPlanSegmentDriverRow {
	rows := make([]projectTransportPlanSegmentDriverRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := projectTransportPlanSegmentDriverRow{
			ID:        resource.ID,
			CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
			UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
		}
		if rel, ok := resource.Relationships["project-transport-plan-segment"]; ok && rel.Data != nil {
			row.ProjectTransportPlanSegment = rel.Data.ID
		}
		if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
			row.Driver = rel.Data.ID
		}
		rows = append(rows, row)
	}
	return rows
}

func renderProjectTransportPlanSegmentDriversTable(cmd *cobra.Command, rows []projectTransportPlanSegmentDriverRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan segment drivers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSEGMENT\tDRIVER\tCREATED AT\tUPDATED AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.ProjectTransportPlanSegment,
			row.Driver,
			row.CreatedAt,
			row.UpdatedAt,
		)
	}
	return writer.Flush()
}
