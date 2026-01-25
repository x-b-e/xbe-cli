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

type jobProductionPlanJobSiteLocationEstimatesListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	JobProductionPlan string
}

type jobProductionPlanJobSiteLocationEstimateRow struct {
	ID                string `json:"id"`
	JobProductionPlan string `json:"job_production_plan_id,omitempty"`
	EstimateCount     int    `json:"estimate_count"`
	Estimates         any    `json:"estimates,omitempty"`
}

func newJobProductionPlanJobSiteLocationEstimatesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan job site location estimates",
		Long: `List job production plan job site location estimates.

Output Columns:
  ID            Estimate ID
  JOB PLAN      Job production plan ID
  ESTIMATES     Estimate count

Filters:
  --job-production-plan  Filter by job production plan ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List job site location estimates
  xbe view job-production-plan-job-site-location-estimates list

  # Filter by job production plan
  xbe view job-production-plan-job-site-location-estimates list --job-production-plan 123

  # Output as JSON
  xbe view job-production-plan-job-site-location-estimates list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanJobSiteLocationEstimatesList,
	}
	initJobProductionPlanJobSiteLocationEstimatesListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanJobSiteLocationEstimatesCmd.AddCommand(newJobProductionPlanJobSiteLocationEstimatesListCmd())
}

func initJobProductionPlanJobSiteLocationEstimatesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanJobSiteLocationEstimatesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanJobSiteLocationEstimatesListOptions(cmd)
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
	query.Set("fields[job-production-plan-job-site-location-estimates]", "estimates,job-production-plan")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlan)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-job-site-location-estimates", query)
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

	rows := buildJobProductionPlanJobSiteLocationEstimateRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanJobSiteLocationEstimatesTable(cmd, rows)
}

func parseJobProductionPlanJobSiteLocationEstimatesListOptions(cmd *cobra.Command) (jobProductionPlanJobSiteLocationEstimatesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanJobSiteLocationEstimatesListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		JobProductionPlan: jobProductionPlan,
	}, nil
}

func buildJobProductionPlanJobSiteLocationEstimateRows(resp jsonAPIResponse) []jobProductionPlanJobSiteLocationEstimateRow {
	rows := make([]jobProductionPlanJobSiteLocationEstimateRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildJobProductionPlanJobSiteLocationEstimateRow(resource)
		row.EstimateCount = estimateCountValue(resource.Attributes["estimates"])
		rows = append(rows, row)
	}
	return rows
}

func buildJobProductionPlanJobSiteLocationEstimateRow(resource jsonAPIResource) jobProductionPlanJobSiteLocationEstimateRow {
	row := jobProductionPlanJobSiteLocationEstimateRow{
		ID: resource.ID,
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlan = rel.Data.ID
	}

	return row
}

func buildJobProductionPlanJobSiteLocationEstimateRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanJobSiteLocationEstimateRow {
	row := buildJobProductionPlanJobSiteLocationEstimateRow(resp.Data)
	row.Estimates = resp.Data.Attributes["estimates"]
	row.EstimateCount = estimateCountValue(row.Estimates)
	return row
}

func renderJobProductionPlanJobSiteLocationEstimatesTable(cmd *cobra.Command, rows []jobProductionPlanJobSiteLocationEstimateRow) error {
	out := cmd.OutOrStdout()
	writer := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)

	fmt.Fprintln(writer, "ID\tJOB PLAN\tESTIMATES")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%d\n", row.ID, row.JobProductionPlan, row.EstimateCount)
	}

	return writer.Flush()
}

func estimateCountValue(value any) int {
	switch typed := value.(type) {
	case []any:
		return len(typed)
	default:
		return 0
	}
}
