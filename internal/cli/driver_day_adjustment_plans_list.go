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

type driverDayAdjustmentPlansListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
	Trucker string
}

type driverDayAdjustmentPlanRow struct {
	ID               string `json:"id"`
	TruckerID        string `json:"trucker_id,omitempty"`
	Content          string `json:"content,omitempty"`
	StartAt          string `json:"start_at,omitempty"`
	StartAtEffective string `json:"start_at_effective,omitempty"`
}

func newDriverDayAdjustmentPlansListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List driver day adjustment plans",
		Long: `List driver day adjustment plans.

Output Columns:
  ID            Plan identifier
  TRUCKER       Trucker ID
  START AT      Plan start timestamp
  EFFECTIVE AT  Effective start timestamp
  CONTENT       Plan content (truncated)

Filters:
  --trucker     Filter by trucker ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List driver day adjustment plans
  xbe view driver-day-adjustment-plans list

  # Filter by trucker
  xbe view driver-day-adjustment-plans list --trucker 123

  # Output as JSON
  xbe view driver-day-adjustment-plans list --json`,
		Args: cobra.NoArgs,
		RunE: runDriverDayAdjustmentPlansList,
	}
	initDriverDayAdjustmentPlansListFlags(cmd)
	return cmd
}

func init() {
	driverDayAdjustmentPlansCmd.AddCommand(newDriverDayAdjustmentPlansListCmd())
}

func initDriverDayAdjustmentPlansListFlags(cmd *cobra.Command) {
	cmd.Flags().String("trucker", "", "Filter by trucker ID")

	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Limit results")
	cmd.Flags().Int("offset", 0, "Offset results")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverDayAdjustmentPlansList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDriverDayAdjustmentPlansListOptions(cmd)
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
	query.Set("include", "trucker")
	query.Set("fields[driver-day-adjustment-plans]", "content,start-at,start-at-effective,trucker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)

	body, _, err := client.Get(cmd.Context(), "/v1/driver-day-adjustment-plans", query)
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

	rows := buildDriverDayAdjustmentPlanRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDriverDayAdjustmentPlansTable(cmd, rows)
}

func parseDriverDayAdjustmentPlansListOptions(cmd *cobra.Command) (driverDayAdjustmentPlansListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	trucker, _ := cmd.Flags().GetString("trucker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverDayAdjustmentPlansListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
		Trucker: trucker,
	}, nil
}

func buildDriverDayAdjustmentPlanRows(resp jsonAPIResponse) []driverDayAdjustmentPlanRow {
	rows := make([]driverDayAdjustmentPlanRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, driverDayAdjustmentPlanRowFromResource(resource))
	}
	return rows
}

func driverDayAdjustmentPlanRowFromSingle(resp jsonAPISingleResponse) driverDayAdjustmentPlanRow {
	return driverDayAdjustmentPlanRowFromResource(resp.Data)
}

func driverDayAdjustmentPlanRowFromResource(resource jsonAPIResource) driverDayAdjustmentPlanRow {
	row := driverDayAdjustmentPlanRow{
		ID:               resource.ID,
		Content:          strings.TrimSpace(stringAttr(resource.Attributes, "content")),
		StartAt:          formatDateTime(stringAttr(resource.Attributes, "start-at")),
		StartAtEffective: formatDateTime(stringAttr(resource.Attributes, "start-at-effective")),
	}

	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
	}

	return row
}

func renderDriverDayAdjustmentPlansTable(cmd *cobra.Command, rows []driverDayAdjustmentPlanRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No driver day adjustment plans found.")
		return nil
	}

	const contentMax = 40

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTRUCKER\tSTART AT\tEFFECTIVE AT\tCONTENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TruckerID,
			row.StartAt,
			row.StartAtEffective,
			truncateString(row.Content, contentMax),
		)
	}
	return writer.Flush()
}
