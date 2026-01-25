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

type jobProductionPlanRecapsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Plan         string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
}

type jobProductionPlanRecapRow struct {
	ID        string `json:"id"`
	PlanID    string `json:"plan_id,omitempty"`
	Plan      string `json:"plan,omitempty"`
	Markdown  string `json:"markdown,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func newJobProductionPlanRecapsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan recaps",
		Long: `List job production plan recaps.

Output Columns:
  ID        Recap identifier
  PLAN      Job production plan name/number
  CREATED   Created timestamp
  RECAP     Recap markdown preview

Filters:
  --plan             Filter by job production plan ID
  --created-at-min   Filter by created-at on/after (ISO 8601)
  --created-at-max   Filter by created-at on/before (ISO 8601)
  --updated-at-min   Filter by updated-at on/after (ISO 8601)
  --updated-at-max   Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List recaps
  xbe view job-production-plan-recaps list

  # Filter by job production plan
  xbe view job-production-plan-recaps list --plan 123

  # Output as JSON
  xbe view job-production-plan-recaps list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanRecapsList,
	}
	initJobProductionPlanRecapsListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanRecapsCmd.AddCommand(newJobProductionPlanRecapsListCmd())
}

func initJobProductionPlanRecapsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("plan", "", "Filter by job production plan ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanRecapsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanRecapsListOptions(cmd)
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
	query.Set("fields[job-production-plan-recaps]", "created-at,updated-at,plan,markdown")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("include", "plan")
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "-created-at")
	}

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[plan]", opts.Plan)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-recaps", query)
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

	rows := buildJobProductionPlanRecapRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanRecapsTable(cmd, rows)
}

func parseJobProductionPlanRecapsListOptions(cmd *cobra.Command) (jobProductionPlanRecapsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	plan, _ := cmd.Flags().GetString("plan")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanRecapsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Plan:         plan,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
	}, nil
}

func buildJobProductionPlanRecapRows(resp jsonAPIResponse) []jobProductionPlanRecapRow {
	rows := make([]jobProductionPlanRecapRow, 0, len(resp.Data))
	included := make(map[string]jsonAPIResource)
	for _, resource := range resp.Included {
		included[resourceKey(resource.Type, resource.ID)] = resource
	}

	for _, resource := range resp.Data {
		rows = append(rows, buildJobProductionPlanRecapRow(resource, included))
	}
	return rows
}

func buildJobProductionPlanRecapRow(resource jsonAPIResource, included map[string]jsonAPIResource) jobProductionPlanRecapRow {
	attrs := resource.Attributes
	row := jobProductionPlanRecapRow{
		ID:        resource.ID,
		Markdown:  strings.TrimSpace(stringAttr(attrs, "markdown")),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	planType := ""
	if rel, ok := resource.Relationships["plan"]; ok && rel.Data != nil {
		row.PlanID = rel.Data.ID
		planType = rel.Data.Type
	}

	if len(included) == 0 {
		return row
	}

	if row.PlanID != "" && planType != "" {
		if plan, ok := included[resourceKey(planType, row.PlanID)]; ok {
			jobNumber := strings.TrimSpace(stringAttr(plan.Attributes, "job-number"))
			jobName := strings.TrimSpace(stringAttr(plan.Attributes, "job-name"))
			if jobNumber != "" && jobName != "" {
				row.Plan = fmt.Sprintf("%s - %s", jobNumber, jobName)
			} else {
				row.Plan = firstNonEmpty(jobNumber, jobName)
			}
		}
	}

	return row
}

func buildJobProductionPlanRecapRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanRecapRow {
	included := make(map[string]jsonAPIResource)
	for _, resource := range resp.Included {
		included[resourceKey(resource.Type, resource.ID)] = resource
	}

	return buildJobProductionPlanRecapRow(resp.Data, included)
}

func renderJobProductionPlanRecapsTable(cmd *cobra.Command, rows []jobProductionPlanRecapRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan recaps found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPLAN\tCREATED\tRECAP")
	for _, row := range rows {
		preview := truncateString(stripMarkdown(row.Markdown), 80)
		created := firstNonEmpty(row.CreatedAt, row.UpdatedAt)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(firstNonEmpty(row.Plan, row.PlanID), 32),
			truncateString(created, 20),
			preview,
		)
	}
	return writer.Flush()
}
