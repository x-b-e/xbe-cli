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

type jobProductionPlanCostCodesListOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	NoAuth                        bool
	Limit                         int
	Offset                        int
	Sort                          string
	JobProductionPlanID           string
	CostCodeID                    string
	ProjectResourceClassification string
	Code                          string
	Query                         string
	Description                   string
}

type jobProductionPlanCostCodeRow struct {
	ID                            string `json:"id"`
	JobProductionPlanID           string `json:"job_production_plan_id,omitempty"`
	CostCodeID                    string `json:"cost_code_id,omitempty"`
	ProjectResourceClassification string `json:"project_resource_classification_id,omitempty"`
}

func newJobProductionPlanCostCodesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan cost codes",
		Long: `List job production plan cost codes.

Output Columns:
  ID               Job production plan cost code identifier
  PLAN             Job production plan ID
  COST CODE        Cost code ID
  RESOURCE CLASS   Project resource classification ID

Filters:
  --job-production-plan           Filter by job production plan ID
  --cost-code                     Filter by cost code ID
  --project-resource-classification Filter by project resource classification ID
  --code                          Filter by cost code value
  --query                         Search by cost code code or description
  --description                   Filter by cost code description

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List job production plan cost codes
  xbe view job-production-plan-cost-codes list

  # Filter by job production plan
  xbe view job-production-plan-cost-codes list --job-production-plan 123

  # Filter by cost code
  xbe view job-production-plan-cost-codes list --cost-code 456

  # Search by cost code
  xbe view job-production-plan-cost-codes list --query "labor"

  # Output as JSON
  xbe view job-production-plan-cost-codes list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanCostCodesList,
	}
	initJobProductionPlanCostCodesListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanCostCodesCmd.AddCommand(newJobProductionPlanCostCodesListCmd())
}

func initJobProductionPlanCostCodesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("cost-code", "", "Filter by cost code ID")
	cmd.Flags().String("project-resource-classification", "", "Filter by project resource classification ID")
	cmd.Flags().String("code", "", "Filter by cost code value")
	cmd.Flags().String("query", "", "Search by cost code code or description")
	cmd.Flags().String("description", "", "Filter by cost code description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanCostCodesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanCostCodesListOptions(cmd)
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
	query.Set("include", "job-production-plan,cost-code,project-resource-classification")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[job_production_plan]", opts.JobProductionPlanID)
	setFilterIfPresent(query, "filter[cost_code]", opts.CostCodeID)
	setFilterIfPresent(query, "filter[project_resource_classification]", opts.ProjectResourceClassification)
	setFilterIfPresent(query, "filter[code]", opts.Code)
	setFilterIfPresent(query, "filter[q]", opts.Query)
	setFilterIfPresent(query, "filter[description]", opts.Description)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-cost-codes", query)
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

	rows := buildJobProductionPlanCostCodeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanCostCodesTable(cmd, rows)
}

func parseJobProductionPlanCostCodesListOptions(cmd *cobra.Command) (jobProductionPlanCostCodesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	costCodeID, _ := cmd.Flags().GetString("cost-code")
	projectResourceClassification, _ := cmd.Flags().GetString("project-resource-classification")
	code, _ := cmd.Flags().GetString("code")
	queryStr, _ := cmd.Flags().GetString("query")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanCostCodesListOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		NoAuth:                        noAuth,
		Limit:                         limit,
		Offset:                        offset,
		Sort:                          sort,
		JobProductionPlanID:           jobProductionPlanID,
		CostCodeID:                    costCodeID,
		ProjectResourceClassification: projectResourceClassification,
		Code:                          code,
		Query:                         queryStr,
		Description:                   description,
	}, nil
}

func buildJobProductionPlanCostCodeRows(resp jsonAPIResponse) []jobProductionPlanCostCodeRow {
	rows := make([]jobProductionPlanCostCodeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := jobProductionPlanCostCodeRow{
			ID: resource.ID,
		}

		if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlanID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["cost-code"]; ok && rel.Data != nil {
			row.CostCodeID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["project-resource-classification"]; ok && rel.Data != nil {
			row.ProjectResourceClassification = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderJobProductionPlanCostCodesTable(cmd *cobra.Command, rows []jobProductionPlanCostCodeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan cost codes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPLAN\tCOST CODE\tRESOURCE CLASS")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobProductionPlanID,
			row.CostCodeID,
			row.ProjectResourceClassification,
		)
	}
	return writer.Flush()
}
