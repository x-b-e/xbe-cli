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

type lineupScenariosListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	Sort      string
	Broker    string
	Date      string
	DateMin   string
	DateMax   string
	Window    string
	Generator string
}

type lineupScenarioRow struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	BrokerID    string `json:"broker_id,omitempty"`
	CustomerID  string `json:"customer_id,omitempty"`
	Date        string `json:"date,omitempty"`
	Window      string `json:"window,omitempty"`
	GeneratorID string `json:"generator_id,omitempty"`
}

func newLineupScenariosListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List lineup scenarios",
		Long: `List lineup scenarios.

Output Columns:
  ID         Lineup scenario identifier
  NAME       Scenario name
  BROKER     Broker ID
  CUSTOMER   Customer ID (if set)
  DATE       Scenario date
  WINDOW     Scenario window (day/night)
  GENERATOR  Generator ID (if set)

Filters:
  --broker     Filter by broker ID
  --date       Filter by date (YYYY-MM-DD)
  --date-min   Filter by date on/after (YYYY-MM-DD)
  --date-max   Filter by date on/before (YYYY-MM-DD)
  --window     Filter by window (day or night)
  --generator  Filter by generator ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List lineup scenarios
  xbe view lineup-scenarios list

  # Filter by broker/date/window
  xbe view lineup-scenarios list --broker 123 --date 2026-01-23 --window day

  # Filter by generator
  xbe view lineup-scenarios list --generator 456

  # Output as JSON
  xbe view lineup-scenarios list --json`,
		Args: cobra.NoArgs,
		RunE: runLineupScenariosList,
	}
	initLineupScenariosListFlags(cmd)
	return cmd
}

func init() {
	lineupScenariosCmd.AddCommand(newLineupScenariosListCmd())
}

func initLineupScenariosListFlags(cmd *cobra.Command) {
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
	cmd.Flags().String("generator", "", "Filter by generator ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupScenariosList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLineupScenariosListOptions(cmd)
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
	query.Set("fields[lineup-scenarios]", "name,date,window")

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
	setFilterIfPresent(query, "filter[generator]", opts.Generator)

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-scenarios", query)
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

	rows := buildLineupScenarioRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLineupScenariosTable(cmd, rows)
}

func parseLineupScenariosListOptions(cmd *cobra.Command) (lineupScenariosListOptions, error) {
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
	generator, _ := cmd.Flags().GetString("generator")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupScenariosListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		Sort:      sort,
		Broker:    broker,
		Date:      date,
		DateMin:   dateMin,
		DateMax:   dateMax,
		Window:    window,
		Generator: generator,
	}, nil
}

func buildLineupScenarioRows(resp jsonAPIResponse) []lineupScenarioRow {
	rows := make([]lineupScenarioRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildLineupScenarioRow(resource))
	}
	return rows
}

func buildLineupScenarioRow(resource jsonAPIResource) lineupScenarioRow {
	attrs := resource.Attributes
	row := lineupScenarioRow{
		ID:     resource.ID,
		Name:   strings.TrimSpace(stringAttr(attrs, "name")),
		Date:   formatDate(stringAttr(attrs, "date")),
		Window: stringAttr(attrs, "window"),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
		row.CustomerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["generator"]; ok && rel.Data != nil {
		row.GeneratorID = rel.Data.ID
	}

	return row
}

func renderLineupScenariosTable(cmd *cobra.Command, rows []lineupScenarioRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No lineup scenarios found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tBROKER\tCUSTOMER\tDATE\tWINDOW\tGENERATOR")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 32),
			row.BrokerID,
			row.CustomerID,
			row.Date,
			row.Window,
			row.GeneratorID,
		)
	}
	return writer.Flush()
}
