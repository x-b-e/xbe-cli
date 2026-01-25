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

type lineupScenarioGeneratorsListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Sort           string
	Broker         string
	Date           string
	DateMin        string
	DateMax        string
	Window         string
	CompletedAtMin string
	CompletedAtMax string
}

type lineupScenarioGeneratorRow struct {
	ID          string `json:"id"`
	BrokerID    string `json:"broker_id,omitempty"`
	CustomerID  string `json:"customer_id,omitempty"`
	Date        string `json:"date,omitempty"`
	Window      string `json:"window,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
}

func newLineupScenarioGeneratorsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List lineup scenario generators",
		Long: `List lineup scenario generators.

Output Columns:
  ID         Generator identifier
  BROKER     Broker ID
  CUSTOMER   Customer ID (if set)
  DATE       Scenario date
  WINDOW     Scenario window (day/night)
  COMPLETED  Completed timestamp

Filters:
  --broker            Filter by broker ID
  --date              Filter by date (YYYY-MM-DD)
  --date-min          Filter by date on/after (YYYY-MM-DD)
  --date-max          Filter by date on/before (YYYY-MM-DD)
  --window            Filter by window (day or night)
  --completed-at-min  Filter by completion on/after (ISO 8601)
  --completed-at-max  Filter by completion on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List generators
  xbe view lineup-scenario-generators list

  # Filter by broker/date/window
  xbe view lineup-scenario-generators list --broker 123 --date 2026-01-23 --window day

  # Output JSON
  xbe view lineup-scenario-generators list --json`,
		Args: cobra.NoArgs,
		RunE: runLineupScenarioGeneratorsList,
	}
	initLineupScenarioGeneratorsListFlags(cmd)
	return cmd
}

func init() {
	lineupScenarioGeneratorsCmd.AddCommand(newLineupScenarioGeneratorsListCmd())
}

func initLineupScenarioGeneratorsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("date", "", "Filter by date (YYYY-MM-DD)")
	cmd.Flags().String("date-min", "", "Filter by date on/after (YYYY-MM-DD)")
	cmd.Flags().String("date-max", "", "Filter by date on/before (YYYY-MM-DD)")
	cmd.Flags().String("window", "", "Filter by window (day or night)")
	cmd.Flags().String("completed-at-min", "", "Filter by completion on/after (ISO 8601)")
	cmd.Flags().String("completed-at-max", "", "Filter by completion on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupScenarioGeneratorsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLineupScenarioGeneratorsListOptions(cmd)
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
	query.Set("fields[lineup-scenario-generators]", "date,window,completed-at")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[date]", opts.Date)
	setFilterIfPresent(query, "filter[date-min]", opts.DateMin)
	setFilterIfPresent(query, "filter[date-max]", opts.DateMax)
	setFilterIfPresent(query, "filter[window]", opts.Window)
	setFilterIfPresent(query, "filter[completed-at-min]", opts.CompletedAtMin)
	setFilterIfPresent(query, "filter[completed-at-max]", opts.CompletedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-scenario-generators", query)
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

	rows := buildLineupScenarioGeneratorRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLineupScenarioGeneratorsTable(cmd, rows)
}

func parseLineupScenarioGeneratorsListOptions(cmd *cobra.Command) (lineupScenarioGeneratorsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	date, _ := cmd.Flags().GetString("date")
	dateMin, _ := cmd.Flags().GetString("date-min")
	dateMax, _ := cmd.Flags().GetString("date-max")
	window, _ := cmd.Flags().GetString("window")
	completedAtMin, _ := cmd.Flags().GetString("completed-at-min")
	completedAtMax, _ := cmd.Flags().GetString("completed-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupScenarioGeneratorsListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Sort:           sort,
		Broker:         broker,
		Date:           date,
		DateMin:        dateMin,
		DateMax:        dateMax,
		Window:         window,
		CompletedAtMin: completedAtMin,
		CompletedAtMax: completedAtMax,
	}, nil
}

func buildLineupScenarioGeneratorRows(resp jsonAPIResponse) []lineupScenarioGeneratorRow {
	rows := make([]lineupScenarioGeneratorRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildLineupScenarioGeneratorRow(resource))
	}
	return rows
}

func buildLineupScenarioGeneratorRow(resource jsonAPIResource) lineupScenarioGeneratorRow {
	attrs := resource.Attributes
	row := lineupScenarioGeneratorRow{
		ID:          resource.ID,
		Date:        formatDate(stringAttr(attrs, "date")),
		Window:      stringAttr(attrs, "window"),
		CompletedAt: formatDateTime(stringAttr(attrs, "completed-at")),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
		row.CustomerID = rel.Data.ID
	}

	return row
}

func renderLineupScenarioGeneratorsTable(cmd *cobra.Command, rows []lineupScenarioGeneratorRow) error {
	out := cmd.OutOrStdout()
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "ID\tBROKER\tCUSTOMER\tDATE\tWINDOW\tCOMPLETED")
	for _, row := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.BrokerID,
			row.CustomerID,
			row.Date,
			row.Window,
			row.CompletedAt,
		)
	}

	return w.Flush()
}
