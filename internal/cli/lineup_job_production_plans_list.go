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

type lineupJobProductionPlansListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	LineupID            string
	JobProductionPlanID string
}

type lineupJobProductionPlanRow struct {
	ID                  string `json:"id"`
	LineupID            string `json:"lineup_id,omitempty"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
	IsDeletable         bool   `json:"is_deletable"`
}

func newLineupJobProductionPlansListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List lineup job production plans",
		Long: `List lineup job production plans.

Output Columns:
  ID        Lineup job production plan identifier
  LINEUP    Lineup ID
  PLAN      Job production plan ID
  DELETABLE Whether the plan can be deleted

Filters:
  --lineup                Filter by lineup ID
  --job-production-plan   Filter by job production plan ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List lineup job production plans
  xbe view lineup-job-production-plans list

  # Filter by lineup
  xbe view lineup-job-production-plans list --lineup 123

  # Filter by job production plan
  xbe view lineup-job-production-plans list --job-production-plan 456

  # Output as JSON
  xbe view lineup-job-production-plans list --json`,
		Args: cobra.NoArgs,
		RunE: runLineupJobProductionPlansList,
	}
	initLineupJobProductionPlansListFlags(cmd)
	return cmd
}

func init() {
	lineupJobProductionPlansCmd.AddCommand(newLineupJobProductionPlansListCmd())
}

func initLineupJobProductionPlansListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("lineup", "", "Filter by lineup ID")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupJobProductionPlansList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLineupJobProductionPlansListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[lineup]", opts.LineupID)
	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlanID)

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-job-production-plans", query)
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

	rows := buildLineupJobProductionPlanRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLineupJobProductionPlansTable(cmd, rows)
}

func parseLineupJobProductionPlansListOptions(cmd *cobra.Command) (lineupJobProductionPlansListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	lineupID, _ := cmd.Flags().GetString("lineup")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupJobProductionPlansListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		LineupID:            lineupID,
		JobProductionPlanID: jobProductionPlanID,
	}, nil
}

func buildLineupJobProductionPlanRows(resp jsonAPIResponse) []lineupJobProductionPlanRow {
	rows := make([]lineupJobProductionPlanRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := lineupJobProductionPlanRow{
			ID:          resource.ID,
			IsDeletable: boolAttr(attrs, "is-deletable"),
		}

		if rel, ok := resource.Relationships["lineup"]; ok && rel.Data != nil {
			row.LineupID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlanID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderLineupJobProductionPlansTable(cmd *cobra.Command, rows []lineupJobProductionPlanRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No lineup job production plans found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tLINEUP\tPLAN\tDELETABLE")
	for _, row := range rows {
		deletable := "no"
		if row.IsDeletable {
			deletable = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.LineupID,
			row.JobProductionPlanID,
			deletable,
		)
	}
	return writer.Flush()
}
