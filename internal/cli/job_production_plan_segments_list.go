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

type jobProductionPlanSegmentsListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	JobProductionPlan string
	MaterialSite      string
	MaterialType      string
}

type jobProductionPlanSegmentRow struct {
	ID                            string `json:"id"`
	JobProductionPlanID           string `json:"job_production_plan_id,omitempty"`
	JobProductionPlanSegmentSetID string `json:"job_production_plan_segment_set_id,omitempty"`
	Sequence                      string `json:"sequence,omitempty"`
	Description                   string `json:"description,omitempty"`
	Quantity                      string `json:"quantity,omitempty"`
	MaterialSiteID                string `json:"material_site_id,omitempty"`
	MaterialTypeID                string `json:"material_type_id,omitempty"`
	CostCodeID                    string `json:"cost_code_id,omitempty"`
}

func newJobProductionPlanSegmentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan segments",
		Long: `List job production plan segments with filtering and pagination.

Output Columns:
  ID            Segment identifier
  JOB_PLAN      Job production plan ID
  SEQ           Segment sequence
  DESC          Segment description
  QTY           Planned quantity
  MATERIAL_SITE Material site ID
  MATERIAL_TYPE Material type ID
  COST_CODE     Cost code ID
  SEGMENT_SET   Segment set ID

Filters:
  --job-production-plan  Filter by job production plan ID
  --material-site        Filter by material site ID
  --material-type        Filter by material type ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List segments
  xbe view job-production-plan-segments list

  # Filter by job production plan
  xbe view job-production-plan-segments list --job-production-plan 123

  # Filter by material site
  xbe view job-production-plan-segments list --material-site 456

  # Output as JSON
  xbe view job-production-plan-segments list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanSegmentsList,
	}
	initJobProductionPlanSegmentsListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanSegmentsCmd.AddCommand(newJobProductionPlanSegmentsListCmd())
}

func initJobProductionPlanSegmentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("material-site", "", "Filter by material site ID")
	cmd.Flags().String("material-type", "", "Filter by material type ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanSegmentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanSegmentsListOptions(cmd)
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
	query.Set("fields[job-production-plan-segments]", "description,quantity,sequence,job-production-plan,job-production-plan-segment-set,material-site,material-type,cost-code")

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
	setFilterIfPresent(query, "filter[material-site]", opts.MaterialSite)
	setFilterIfPresent(query, "filter[material-type]", opts.MaterialType)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-segments", query)
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

	rows := buildJobProductionPlanSegmentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanSegmentsTable(cmd, rows)
}

func parseJobProductionPlanSegmentsListOptions(cmd *cobra.Command) (jobProductionPlanSegmentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	materialSite, _ := cmd.Flags().GetString("material-site")
	materialType, _ := cmd.Flags().GetString("material-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanSegmentsListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		JobProductionPlan: jobProductionPlan,
		MaterialSite:      materialSite,
		MaterialType:      materialType,
	}, nil
}

func buildJobProductionPlanSegmentRows(resp jsonAPIResponse) []jobProductionPlanSegmentRow {
	rows := make([]jobProductionPlanSegmentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := jobProductionPlanSegmentRow{
			ID:          resource.ID,
			Sequence:    stringAttr(resource.Attributes, "sequence"),
			Description: stringAttr(resource.Attributes, "description"),
			Quantity:    stringAttr(resource.Attributes, "quantity"),
		}

		row.JobProductionPlanID = relationshipIDFromMap(resource.Relationships, "job-production-plan")
		row.JobProductionPlanSegmentSetID = relationshipIDFromMap(resource.Relationships, "job-production-plan-segment-set")
		row.MaterialSiteID = relationshipIDFromMap(resource.Relationships, "material-site")
		row.MaterialTypeID = relationshipIDFromMap(resource.Relationships, "material-type")
		row.CostCodeID = relationshipIDFromMap(resource.Relationships, "cost-code")

		rows = append(rows, row)
	}
	return rows
}

func renderJobProductionPlanSegmentsTable(cmd *cobra.Command, rows []jobProductionPlanSegmentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan segments found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tJOB_PLAN\tSEQ\tDESC\tQTY\tMATERIAL_SITE\tMATERIAL_TYPE\tCOST_CODE\tSEGMENT_SET")
	for _, row := range rows {
		desc := truncateString(row.Description, 30)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobProductionPlanID,
			row.Sequence,
			desc,
			row.Quantity,
			row.MaterialSiteID,
			row.MaterialTypeID,
			row.CostCodeID,
			row.JobProductionPlanSegmentSetID,
		)
	}
	return writer.Flush()
}
