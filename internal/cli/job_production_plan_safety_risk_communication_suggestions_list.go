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

type jobProductionPlanSafetyRiskCommunicationSuggestionsListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	JobProductionPlan string
	CreatedAtMin      string
	CreatedAtMax      string
	UpdatedAtMin      string
	UpdatedAtMax      string
}

type jobProductionPlanSafetyRiskCommunicationSuggestionRow struct {
	ID                  string `json:"id"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
	JobProductionPlan   string `json:"job_production_plan,omitempty"`
	IsAsync             bool   `json:"is_async"`
	IsFulfilled         bool   `json:"is_fulfilled"`
	Suggestion          string `json:"suggestion,omitempty"`
	CreatedAt           string `json:"created_at,omitempty"`
	UpdatedAt           string `json:"updated_at,omitempty"`
}

func newJobProductionPlanSafetyRiskCommunicationSuggestionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List safety risk communication suggestions",
		Long: `List job production plan safety risk communication suggestions.

Output Columns:
  ID                   Suggestion identifier
  JOB PRODUCTION PLAN  Job production plan name/number
  FULFILLED            Whether the suggestion has completed generation
  ASYNC                Whether the suggestion was generated asynchronously
  CREATED              Created timestamp
  SUGGESTION           Suggestion preview

Filters:
  --job-production-plan  Filter by job production plan ID (comma-separated for multiple)
  --created-at-min       Filter by created-at on/after (ISO 8601)
  --created-at-max       Filter by created-at on/before (ISO 8601)
  --updated-at-min       Filter by updated-at on/after (ISO 8601)
  --updated-at-max       Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List suggestions for a job production plan
  xbe view job-production-plan-safety-risk-communication-suggestions list --job-production-plan 123

  # List with a created-at filter
  xbe view job-production-plan-safety-risk-communication-suggestions list --created-at-min 2025-01-01T00:00:00Z

  # Output as JSON
  xbe view job-production-plan-safety-risk-communication-suggestions list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanSafetyRiskCommunicationSuggestionsList,
	}
	initJobProductionPlanSafetyRiskCommunicationSuggestionsListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanSafetyRiskCommunicationSuggestionsCmd.AddCommand(newJobProductionPlanSafetyRiskCommunicationSuggestionsListCmd())
}

func initJobProductionPlanSafetyRiskCommunicationSuggestionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID (comma-separated for multiple)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanSafetyRiskCommunicationSuggestionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanSafetyRiskCommunicationSuggestionsListOptions(cmd)
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
	query.Set("fields[job-production-plan-safety-risk-communication-suggestions]", "created-at,updated-at,job-production-plan,is-async,is-fulfilled,suggestion")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("include", "job-production-plan")
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

	setFilterIfPresent(query, "filter[job_production_plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-safety-risk-communication-suggestions", query)
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

	rows := buildJobProductionPlanSafetyRiskCommunicationSuggestionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanSafetyRiskCommunicationSuggestionsTable(cmd, rows)
}

func parseJobProductionPlanSafetyRiskCommunicationSuggestionsListOptions(cmd *cobra.Command) (jobProductionPlanSafetyRiskCommunicationSuggestionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanSafetyRiskCommunicationSuggestionsListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		JobProductionPlan: jobProductionPlan,
		CreatedAtMin:      createdAtMin,
		CreatedAtMax:      createdAtMax,
		UpdatedAtMin:      updatedAtMin,
		UpdatedAtMax:      updatedAtMax,
	}, nil
}

func buildJobProductionPlanSafetyRiskCommunicationSuggestionRows(resp jsonAPIResponse) []jobProductionPlanSafetyRiskCommunicationSuggestionRow {
	rows := make([]jobProductionPlanSafetyRiskCommunicationSuggestionRow, 0, len(resp.Data))
	included := make(map[string]jsonAPIResource)
	for _, resource := range resp.Included {
		included[resourceKey(resource.Type, resource.ID)] = resource
	}

	for _, resource := range resp.Data {
		rows = append(rows, buildJobProductionPlanSafetyRiskCommunicationSuggestionRow(resource, included))
	}
	return rows
}

func buildJobProductionPlanSafetyRiskCommunicationSuggestionRow(resource jsonAPIResource, included map[string]jsonAPIResource) jobProductionPlanSafetyRiskCommunicationSuggestionRow {
	attrs := resource.Attributes
	row := jobProductionPlanSafetyRiskCommunicationSuggestionRow{
		ID:          resource.ID,
		IsAsync:     boolAttr(attrs, "is-async"),
		IsFulfilled: boolAttr(attrs, "is-fulfilled"),
		Suggestion:  strings.TrimSpace(stringAttr(attrs, "suggestion")),
		CreatedAt:   formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:   formatDateTime(stringAttr(attrs, "updated-at")),
	}

	jppType := ""
	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
		jppType = rel.Data.Type
	}

	if row.JobProductionPlanID != "" && jppType != "" {
		if jpp, ok := included[resourceKey(jppType, row.JobProductionPlanID)]; ok {
			jobNumber := strings.TrimSpace(stringAttr(jpp.Attributes, "job-number"))
			jobName := strings.TrimSpace(stringAttr(jpp.Attributes, "job-name"))
			if jobNumber != "" && jobName != "" {
				row.JobProductionPlan = fmt.Sprintf("%s - %s", jobNumber, jobName)
			} else {
				row.JobProductionPlan = firstNonEmpty(jobNumber, jobName)
			}
		}
	}

	return row
}

func renderJobProductionPlanSafetyRiskCommunicationSuggestionsTable(cmd *cobra.Command, rows []jobProductionPlanSafetyRiskCommunicationSuggestionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan safety risk communication suggestions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tJOB PRODUCTION PLAN\tFULFILLED\tASYNC\tCREATED\tSUGGESTION")

	for _, row := range rows {
		preview := suggestionPreview(row.Suggestion, row.IsFulfilled)
		fmt.Fprintf(writer, "%s\t%s\t%t\t%t\t%s\t%s\n",
			row.ID,
			truncateString(firstNonEmpty(row.JobProductionPlan, row.JobProductionPlanID), 32),
			row.IsFulfilled,
			row.IsAsync,
			truncateString(row.CreatedAt, 20),
			truncateString(preview, 60),
		)
	}

	return writer.Flush()
}

func suggestionPreview(value string, isFulfilled bool) string {
	preview := strings.TrimSpace(value)
	if preview == "" && !isFulfilled {
		return "Pending"
	}
	preview = stripMarkdown(preview)
	preview = strings.Join(strings.Fields(preview), " ")
	return preview
}
