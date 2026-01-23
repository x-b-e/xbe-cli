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

type jobProductionPlanDisplayUnitOfMeasuresListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	JobProductionPlan string
	UnitOfMeasure     string
}

type jobProductionPlanDisplayUnitOfMeasureRow struct {
	ID                 string `json:"id"`
	JobProductionPlan  string `json:"job_production_plan_id,omitempty"`
	UnitOfMeasure      string `json:"unit_of_measure_id,omitempty"`
	Importance         int    `json:"importance"`
	ImportancePosition int    `json:"importance_position"`
}

func newJobProductionPlanDisplayUnitOfMeasuresListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan display unit of measures",
		Long: `List job production plan display unit of measures.

Output Columns:
  ID            Display unit of measure ID
  JOB PLAN      Job production plan ID
  UNIT          Unit of measure ID
  IMPORTANCE    Importance (computed)
  POSITION      Importance position (0-based)

Filters:
  --job-production-plan  Filter by job production plan ID
  --unit-of-measure      Filter by unit of measure ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List display unit of measures
  xbe view job-production-plan-display-unit-of-measures list

  # Filter by job production plan
  xbe view job-production-plan-display-unit-of-measures list --job-production-plan 123

  # Filter by unit of measure
  xbe view job-production-plan-display-unit-of-measures list --unit-of-measure 456

  # Output as JSON
  xbe view job-production-plan-display-unit-of-measures list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanDisplayUnitOfMeasuresList,
	}
	initJobProductionPlanDisplayUnitOfMeasuresListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanDisplayUnitOfMeasuresCmd.AddCommand(newJobProductionPlanDisplayUnitOfMeasuresListCmd())
}

func initJobProductionPlanDisplayUnitOfMeasuresListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("unit-of-measure", "", "Filter by unit of measure ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanDisplayUnitOfMeasuresList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanDisplayUnitOfMeasuresListOptions(cmd)
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
	query.Set("fields[job-production-plan-display-unit-of-measures]", "importance,importance-position,job-production-plan,unit-of-measure")

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
	setFilterIfPresent(query, "filter[unit-of-measure]", opts.UnitOfMeasure)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-display-unit-of-measures", query)
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

	rows := buildJobProductionPlanDisplayUnitOfMeasureRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanDisplayUnitOfMeasuresTable(cmd, rows)
}

func parseJobProductionPlanDisplayUnitOfMeasuresListOptions(cmd *cobra.Command) (jobProductionPlanDisplayUnitOfMeasuresListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanDisplayUnitOfMeasuresListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		JobProductionPlan: jobProductionPlan,
		UnitOfMeasure:     unitOfMeasure,
	}, nil
}

func buildJobProductionPlanDisplayUnitOfMeasureRows(resp jsonAPIResponse) []jobProductionPlanDisplayUnitOfMeasureRow {
	rows := make([]jobProductionPlanDisplayUnitOfMeasureRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildJobProductionPlanDisplayUnitOfMeasureRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildJobProductionPlanDisplayUnitOfMeasureRow(resource jsonAPIResource) jobProductionPlanDisplayUnitOfMeasureRow {
	row := jobProductionPlanDisplayUnitOfMeasureRow{
		ID:                 resource.ID,
		Importance:         intAttr(resource.Attributes, "importance"),
		ImportancePosition: intAttr(resource.Attributes, "importance-position"),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlan = rel.Data.ID
	}
	if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		row.UnitOfMeasure = rel.Data.ID
	}

	return row
}

func buildJobProductionPlanDisplayUnitOfMeasureRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanDisplayUnitOfMeasureRow {
	return buildJobProductionPlanDisplayUnitOfMeasureRow(resp.Data)
}

func renderJobProductionPlanDisplayUnitOfMeasuresTable(cmd *cobra.Command, rows []jobProductionPlanDisplayUnitOfMeasureRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan display unit of measures found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tJOB PLAN\tUNIT\tIMPORTANCE\tPOSITION")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%d\t%d\n",
			row.ID,
			row.JobProductionPlan,
			row.UnitOfMeasure,
			row.Importance,
			row.ImportancePosition,
		)
	}
	return writer.Flush()
}
