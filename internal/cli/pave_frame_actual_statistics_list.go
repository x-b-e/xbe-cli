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

type paveFrameActualStatisticsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	CreatedAtMin string
	CreatedAtMax string
	IsCreatedAt  string
	UpdatedAtMin string
	UpdatedAtMax string
	IsUpdatedAt  string
}

type paveFrameActualStatisticRow struct {
	ID        string `json:"id"`
	Name      string `json:"name,omitempty"`
	Latitude  string `json:"latitude,omitempty"`
	Longitude string `json:"longitude,omitempty"`
	Window    string `json:"window,omitempty"`
	AggLevel  string `json:"agg_level,omitempty"`
	DateMin   string `json:"date_min,omitempty"`
	DateMax   string `json:"date_max,omitempty"`
}

func newPaveFrameActualStatisticsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pave frame actual statistics",
		Long: `List pave frame actual statistics with filtering and pagination.

Output Columns:
  ID        Statistic identifier
  LOCATION  Name or latitude/longitude
  WINDOW    Paving window (day/night)
  AGG       Aggregation level
  DATE MIN  Optional start date filter
  DATE MAX  Optional end date filter

Filters:
  --created-at-min Filter by created-at on/after (ISO 8601)
  --created-at-max Filter by created-at on/before (ISO 8601)
  --is-created-at  Filter by has created-at (true/false)
  --updated-at-min Filter by updated-at on/after (ISO 8601)
  --updated-at-max Filter by updated-at on/before (ISO 8601)
  --is-updated-at  Filter by has updated-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List statistics
  xbe view pave-frame-actual-statistics list

  # Filter by created-at
  xbe view pave-frame-actual-statistics list --created-at-min 2025-01-01T00:00:00Z

  # Output as JSON
  xbe view pave-frame-actual-statistics list --json`,
		Args: cobra.NoArgs,
		RunE: runPaveFrameActualStatisticsList,
	}
	initPaveFrameActualStatisticsListFlags(cmd)
	return cmd
}

func init() {
	paveFrameActualStatisticsCmd.AddCommand(newPaveFrameActualStatisticsListCmd())
}

func initPaveFrameActualStatisticsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPaveFrameActualStatisticsList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePaveFrameActualStatisticsListOptions(cmd)
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
	query.Set("fields[pave-frame-actual-statistics]", "name,latitude,longitude,window,agg-level,date-min,date-max")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/pave-frame-actual-statistics", query)
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

	rows := buildPaveFrameActualStatisticRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPaveFrameActualStatisticsTable(cmd, rows)
}

func parsePaveFrameActualStatisticsListOptions(cmd *cobra.Command) (paveFrameActualStatisticsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return paveFrameActualStatisticsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		IsCreatedAt:  isCreatedAt,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		IsUpdatedAt:  isUpdatedAt,
	}, nil
}

func buildPaveFrameActualStatisticRows(resp jsonAPIResponse) []paveFrameActualStatisticRow {
	rows := make([]paveFrameActualStatisticRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildPaveFrameActualStatisticRow(resource))
	}
	return rows
}

func buildPaveFrameActualStatisticRow(resource jsonAPIResource) paveFrameActualStatisticRow {
	attrs := resource.Attributes
	return paveFrameActualStatisticRow{
		ID:        resource.ID,
		Name:      stringAttr(attrs, "name"),
		Latitude:  stringAttr(attrs, "latitude"),
		Longitude: stringAttr(attrs, "longitude"),
		Window:    stringAttr(attrs, "window"),
		AggLevel:  stringAttr(attrs, "agg-level"),
		DateMin:   stringAttr(attrs, "date-min"),
		DateMax:   stringAttr(attrs, "date-max"),
	}
}

func renderPaveFrameActualStatisticsTable(cmd *cobra.Command, rows []paveFrameActualStatisticRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No pave frame actual statistics found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tLOCATION\tWINDOW\tAGG\tDATE MIN\tDATE MAX")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			formatPaveFrameActualStatisticLocation(row),
			row.Window,
			row.AggLevel,
			row.DateMin,
			row.DateMax,
		)
	}

	return writer.Flush()
}

func formatPaveFrameActualStatisticLocation(row paveFrameActualStatisticRow) string {
	name := strings.TrimSpace(row.Name)
	if name != "" {
		return name
	}
	lat := strings.TrimSpace(row.Latitude)
	lon := strings.TrimSpace(row.Longitude)
	if lat != "" && lon != "" {
		return fmt.Sprintf("%s,%s", lat, lon)
	}
	if lat != "" {
		return lat
	}
	return lon
}
