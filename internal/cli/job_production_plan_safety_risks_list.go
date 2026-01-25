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

type jobProductionPlanSafetyRisksListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	JobProductionPlan string
}

type jobProductionPlanSafetyRiskRow struct {
	ID                  string `json:"id"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
	JobProductionPlan   string `json:"job_production_plan,omitempty"`
	Description         string `json:"description,omitempty"`
}

func newJobProductionPlanSafetyRisksListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan safety risks",
		Long: `List job production plan safety risks with filtering and pagination.

Output Columns:
  ID          Job production plan safety risk identifier
  JOB PLAN    Job production plan (job number or name)
  DESCRIPTION Safety risk description

Filters:
  --job-production-plan  Filter by job production plan ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List job production plan safety risks
  xbe view job-production-plan-safety-risks list

  # Filter by job production plan
  xbe view job-production-plan-safety-risks list --job-production-plan 123

  # Output as JSON
  xbe view job-production-plan-safety-risks list --json`,
		RunE: runJobProductionPlanSafetyRisksList,
	}
	initJobProductionPlanSafetyRisksListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanSafetyRisksCmd.AddCommand(newJobProductionPlanSafetyRisksListCmd())
}

func initJobProductionPlanSafetyRisksListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanSafetyRisksList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanSafetyRisksListOptions(cmd)
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
	query.Set("fields[job-production-plan-safety-risks]", "description,job-production-plan")
	query.Set("include", "job-production-plan")
	query.Set("fields[job-production-plans]", "job-number,job-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlan)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-safety-risks", query)
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

	rows := buildJobProductionPlanSafetyRiskRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanSafetyRisksTable(cmd, rows)
}

func parseJobProductionPlanSafetyRisksListOptions(cmd *cobra.Command) (jobProductionPlanSafetyRisksListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return jobProductionPlanSafetyRisksListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return jobProductionPlanSafetyRisksListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return jobProductionPlanSafetyRisksListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return jobProductionPlanSafetyRisksListOptions{}, err
	}
	sort, err := cmd.Flags().GetString("sort")
	if err != nil {
		return jobProductionPlanSafetyRisksListOptions{}, err
	}
	jobProductionPlan, err := cmd.Flags().GetString("job-production-plan")
	if err != nil {
		return jobProductionPlanSafetyRisksListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return jobProductionPlanSafetyRisksListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return jobProductionPlanSafetyRisksListOptions{}, err
	}

	return jobProductionPlanSafetyRisksListOptions{
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

func buildJobProductionPlanSafetyRiskRows(resp jsonAPIResponse) []jobProductionPlanSafetyRiskRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]jobProductionPlanSafetyRiskRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildJobProductionPlanSafetyRiskRow(resource, included))
	}
	return rows
}

func jobProductionPlanSafetyRiskRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanSafetyRiskRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildJobProductionPlanSafetyRiskRow(resp.Data, included)
}

func buildJobProductionPlanSafetyRiskRow(resource jsonAPIResource, included map[string]jsonAPIResource) jobProductionPlanSafetyRiskRow {
	attrs := resource.Attributes
	row := jobProductionPlanSafetyRiskRow{
		ID:          resource.ID,
		Description: stringAttr(attrs, "description"),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
		if plan, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.JobProductionPlan = firstNonEmpty(
				stringAttr(plan.Attributes, "job-number"),
				stringAttr(plan.Attributes, "job-name"),
			)
		}
	}

	return row
}

func renderJobProductionPlanSafetyRisksTable(cmd *cobra.Command, rows []jobProductionPlanSafetyRiskRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan safety risks found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tJOB PLAN\tDESCRIPTION")

	for _, row := range rows {
		jobPlan := row.JobProductionPlan
		if jobPlan == "" {
			jobPlan = row.JobProductionPlanID
		}
		description := row.Description
		if description == "" {
			description = "-"
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			jobPlan,
			truncateString(description, 40),
		)
	}

	return writer.Flush()
}
