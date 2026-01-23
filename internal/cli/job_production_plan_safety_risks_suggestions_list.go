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

type jobProductionPlanSafetyRisksSuggestionsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	JobProductionPlanID string
}

type jobProductionPlanSafetyRisksSuggestionRow struct {
	ID                  string `json:"id"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
	IsAsync             bool   `json:"is_async"`
	IsFulfilled         bool   `json:"is_fulfilled"`
	RisksCount          int    `json:"risks_count,omitempty"`
}

func newJobProductionPlanSafetyRisksSuggestionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan safety risks suggestions",
		Long: `List job production plan safety risks suggestions.

Output Columns:
  ID         Suggestion identifier
  PLAN       Job production plan ID
  ASYNC      Whether generation was queued asynchronously
  FULFILLED  Whether suggestions have been generated
  RISKS      Count of generated risks

Filters:
  --job-production-plan  Filter by job production plan ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List safety risks suggestions
  xbe view job-production-plan-safety-risks-suggestions list

  # Filter by job production plan
  xbe view job-production-plan-safety-risks-suggestions list --job-production-plan 123

  # Sort by fulfillment
  xbe view job-production-plan-safety-risks-suggestions list --sort is-fulfilled

  # Output as JSON
  xbe view job-production-plan-safety-risks-suggestions list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanSafetyRisksSuggestionsList,
	}
	initJobProductionPlanSafetyRisksSuggestionsListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanSafetyRisksSuggestionsCmd.AddCommand(newJobProductionPlanSafetyRisksSuggestionsListCmd())
}

func initJobProductionPlanSafetyRisksSuggestionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanSafetyRisksSuggestionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanSafetyRisksSuggestionsListOptions(cmd)
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
	query.Set("fields[job-production-plan-safety-risks-suggestions]", "is-async,is-fulfilled,risks,job-production-plan")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlanID)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-safety-risks-suggestions", query)
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

	rows := buildJobProductionPlanSafetyRisksSuggestionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanSafetyRisksSuggestionsTable(cmd, rows)
}

func parseJobProductionPlanSafetyRisksSuggestionsListOptions(cmd *cobra.Command) (jobProductionPlanSafetyRisksSuggestionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanSafetyRisksSuggestionsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		JobProductionPlanID: jobProductionPlanID,
	}, nil
}

func buildJobProductionPlanSafetyRisksSuggestionRows(resp jsonAPIResponse) []jobProductionPlanSafetyRisksSuggestionRow {
	rows := make([]jobProductionPlanSafetyRisksSuggestionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		risks := stringSliceAttr(attrs, "risks")
		row := jobProductionPlanSafetyRisksSuggestionRow{
			ID:          resource.ID,
			IsAsync:     boolAttr(attrs, "is-async"),
			IsFulfilled: boolAttr(attrs, "is-fulfilled"),
			RisksCount:  len(risks),
		}

		if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlanID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderJobProductionPlanSafetyRisksSuggestionsTable(cmd *cobra.Command, rows []jobProductionPlanSafetyRisksSuggestionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan safety risks suggestions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPLAN\tASYNC\tFULFILLED\tRISKS")
	for _, row := range rows {
		async := ""
		fulfilled := ""
		if row.IsAsync {
			async = "yes"
		}
		if row.IsFulfilled {
			fulfilled = "yes"
		}
		risksCount := ""
		if row.RisksCount > 0 {
			risksCount = strconv.Itoa(row.RisksCount)
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobProductionPlanID,
			async,
			fulfilled,
			risksCount,
		)
	}
	return writer.Flush()
}
