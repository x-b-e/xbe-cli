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

type jobProductionPlanTrailerClassificationsListOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	NoAuth                bool
	Limit                 int
	Offset                int
	Sort                  string
	JobProductionPlan     string
	TrailerClassification string
}

type jobProductionPlanTrailerClassificationRow struct {
	ID                                 string   `json:"id"`
	JobProductionPlan                  string   `json:"job_production_plan_id,omitempty"`
	TrailerClassification              string   `json:"trailer_classification_id,omitempty"`
	TrailerClassificationEquivalentIDs []string `json:"trailer_classification_equivalent_ids,omitempty"`
	GrossWeightLegalLimitLbsExplicit   float64  `json:"gross_weight_legal_limit_lbs_explicit,omitempty"`
	GrossWeightLegalLimitLbs           float64  `json:"gross_weight_legal_limit_lbs,omitempty"`
	ExplicitMaterialTransactionTonsMax float64  `json:"explicit_material_transaction_tons_max,omitempty"`
	MaterialTransactionTonsMax         float64  `json:"material_transaction_tons_max,omitempty"`
}

func newJobProductionPlanTrailerClassificationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan trailer classifications",
		Long: `List job production plan trailer classifications.

Output Columns:
  ID            Job production plan trailer classification ID
  JOB PLAN      Job production plan ID
  TRAILER CLASS Trailer classification ID
  EQUIV IDS     Equivalent trailer classification IDs
  GROSS LIMIT   Gross weight legal limit (lbs)
  TONS MAX      Material transaction tons max

Filters:
  --job-production-plan     Filter by job production plan ID
  --trailer-classification  Filter by trailer classification ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List job production plan trailer classifications
  xbe view job-production-plan-trailer-classifications list

  # Filter by job production plan
  xbe view job-production-plan-trailer-classifications list --job-production-plan 123

  # Filter by trailer classification
  xbe view job-production-plan-trailer-classifications list --trailer-classification 456

  # Output as JSON
  xbe view job-production-plan-trailer-classifications list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanTrailerClassificationsList,
	}
	initJobProductionPlanTrailerClassificationsListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanTrailerClassificationsCmd.AddCommand(newJobProductionPlanTrailerClassificationsListCmd())
}

func initJobProductionPlanTrailerClassificationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("trailer-classification", "", "Filter by trailer classification ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanTrailerClassificationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanTrailerClassificationsListOptions(cmd)
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
	query.Set("fields[job-production-plan-trailer-classifications]", "trailer-classification-equivalent-ids,gross-weight-legal-limit-lbs-explicit,gross-weight-legal-limit-lbs,explicit-material-transaction-tons-max,material-transaction-tons-max,job-production-plan,trailer-classification")

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
	setFilterIfPresent(query, "filter[trailer-classification]", opts.TrailerClassification)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-trailer-classifications", query)
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

	rows := buildJobProductionPlanTrailerClassificationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanTrailerClassificationsTable(cmd, rows)
}

func parseJobProductionPlanTrailerClassificationsListOptions(cmd *cobra.Command) (jobProductionPlanTrailerClassificationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	trailerClassification, _ := cmd.Flags().GetString("trailer-classification")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanTrailerClassificationsListOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		NoAuth:                noAuth,
		Limit:                 limit,
		Offset:                offset,
		Sort:                  sort,
		JobProductionPlan:     jobProductionPlan,
		TrailerClassification: trailerClassification,
	}, nil
}

func buildJobProductionPlanTrailerClassificationRows(resp jsonAPIResponse) []jobProductionPlanTrailerClassificationRow {
	rows := make([]jobProductionPlanTrailerClassificationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildJobProductionPlanTrailerClassificationRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildJobProductionPlanTrailerClassificationRow(resource jsonAPIResource) jobProductionPlanTrailerClassificationRow {
	row := jobProductionPlanTrailerClassificationRow{
		ID:                                 resource.ID,
		TrailerClassificationEquivalentIDs: stringSliceAttr(resource.Attributes, "trailer-classification-equivalent-ids"),
		GrossWeightLegalLimitLbsExplicit:   floatAttr(resource.Attributes, "gross-weight-legal-limit-lbs-explicit"),
		GrossWeightLegalLimitLbs:           floatAttr(resource.Attributes, "gross-weight-legal-limit-lbs"),
		ExplicitMaterialTransactionTonsMax: floatAttr(resource.Attributes, "explicit-material-transaction-tons-max"),
		MaterialTransactionTonsMax:         floatAttr(resource.Attributes, "material-transaction-tons-max"),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlan = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trailer-classification"]; ok && rel.Data != nil {
		row.TrailerClassification = rel.Data.ID
	}

	return row
}

func buildJobProductionPlanTrailerClassificationRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanTrailerClassificationRow {
	return buildJobProductionPlanTrailerClassificationRow(resp.Data)
}

func renderJobProductionPlanTrailerClassificationsTable(cmd *cobra.Command, rows []jobProductionPlanTrailerClassificationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan trailer classifications found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tJOB PLAN\tTRAILER CLASS\tEQUIV IDS\tGROSS LIMIT\tTONS MAX")
	for _, row := range rows {
		equiv := truncateString(strings.Join(row.TrailerClassificationEquivalentIDs, ", "), 25)
		grossLimit := formatOptionalFloat(row.GrossWeightLegalLimitLbs)
		tonsMax := formatOptionalFloat(row.MaterialTransactionTonsMax)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobProductionPlan,
			row.TrailerClassification,
			equiv,
			grossLimit,
			tonsMax,
		)
	}
	return writer.Flush()
}

func formatOptionalFloat(value float64) string {
	if value == 0 {
		return ""
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}
