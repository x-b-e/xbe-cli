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

type jobProductionPlanMaterialTypeQualityControlRequirementsListOptions struct {
	BaseURL                        string
	Token                          string
	JSON                           bool
	NoAuth                         bool
	Limit                          int
	Offset                         int
	Sort                           string
	JobProductionPlanMaterialType  string
	QualityControlClassificationID string
}

type jobProductionPlanMaterialTypeQualityControlRequirementRow struct {
	ID                             string `json:"id"`
	JobProductionPlanMaterialType  string `json:"job_production_plan_material_type_id,omitempty"`
	QualityControlClassificationID string `json:"quality_control_classification_id,omitempty"`
	Note                           string `json:"note,omitempty"`
}

func newJobProductionPlanMaterialTypeQualityControlRequirementsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan material type quality control requirements",
		Long: `List job production plan material type quality control requirements.

Output Columns:
  ID               Requirement identifier
  MATERIAL TYPE    Job production plan material type ID
  QC CLASS         Quality control classification ID
  NOTE             Requirement note

Filters:
  --job-production-plan-material-type  Filter by job production plan material type ID
  --quality-control-classification     Filter by quality control classification ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List requirements
  xbe view job-production-plan-material-type-quality-control-requirements list

  # Filter by job production plan material type
  xbe view job-production-plan-material-type-quality-control-requirements list --job-production-plan-material-type 123

  # Filter by quality control classification
  xbe view job-production-plan-material-type-quality-control-requirements list --quality-control-classification 456

  # Output as JSON
  xbe view job-production-plan-material-type-quality-control-requirements list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanMaterialTypeQualityControlRequirementsList,
	}
	initJobProductionPlanMaterialTypeQualityControlRequirementsListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanMaterialTypeQualityControlRequirementsCmd.AddCommand(newJobProductionPlanMaterialTypeQualityControlRequirementsListCmd())
}

func initJobProductionPlanMaterialTypeQualityControlRequirementsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan-material-type", "", "Filter by job production plan material type ID")
	cmd.Flags().String("quality-control-classification", "", "Filter by quality control classification ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanMaterialTypeQualityControlRequirementsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanMaterialTypeQualityControlRequirementsListOptions(cmd)
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
	query.Set("include", "job-production-plan-material-type,quality-control-classification")
	query.Set("fields[job-production-plan-material-type-quality-control-requirements]", "note,job-production-plan-material-type,quality-control-classification")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[job_production_plan_material_type]", opts.JobProductionPlanMaterialType)
	setFilterIfPresent(query, "filter[quality_control_classification]", opts.QualityControlClassificationID)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-material-type-quality-control-requirements", query)
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

	rows := buildJobProductionPlanMaterialTypeQualityControlRequirementRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanMaterialTypeQualityControlRequirementsTable(cmd, rows)
}

func parseJobProductionPlanMaterialTypeQualityControlRequirementsListOptions(cmd *cobra.Command) (jobProductionPlanMaterialTypeQualityControlRequirementsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlanMaterialType, _ := cmd.Flags().GetString("job-production-plan-material-type")
	qualityControlClassificationID, _ := cmd.Flags().GetString("quality-control-classification")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanMaterialTypeQualityControlRequirementsListOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		NoAuth:                         noAuth,
		Limit:                          limit,
		Offset:                         offset,
		Sort:                           sort,
		JobProductionPlanMaterialType:  jobProductionPlanMaterialType,
		QualityControlClassificationID: qualityControlClassificationID,
	}, nil
}

func buildJobProductionPlanMaterialTypeQualityControlRequirementRows(resp jsonAPIResponse) []jobProductionPlanMaterialTypeQualityControlRequirementRow {
	rows := make([]jobProductionPlanMaterialTypeQualityControlRequirementRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := jobProductionPlanMaterialTypeQualityControlRequirementRow{
			ID:   resource.ID,
			Note: stringAttr(resource.Attributes, "note"),
		}

		if rel, ok := resource.Relationships["job-production-plan-material-type"]; ok && rel.Data != nil {
			row.JobProductionPlanMaterialType = rel.Data.ID
		}
		if rel, ok := resource.Relationships["quality-control-classification"]; ok && rel.Data != nil {
			row.QualityControlClassificationID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderJobProductionPlanMaterialTypeQualityControlRequirementsTable(cmd *cobra.Command, rows []jobProductionPlanMaterialTypeQualityControlRequirementRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan material type quality control requirements found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tMATERIAL TYPE\tQC CLASS\tNOTE")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobProductionPlanMaterialType,
			row.QualityControlClassificationID,
			truncateString(row.Note, 30),
		)
	}
	return writer.Flush()
}
