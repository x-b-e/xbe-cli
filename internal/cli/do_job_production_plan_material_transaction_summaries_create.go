package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doJobProductionPlanMaterialTransactionSummariesCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	JobProductionPlan string
}

type jobProductionPlanMaterialTransactionSummaryItem struct {
	MaterialTypeID                 string `json:"material_type_id,omitempty"`
	MaterialTypeFullyQualifiedName string `json:"material_type_fully_qualified_name,omitempty"`
	Tons                           string `json:"tons,omitempty"`
}

type jobProductionPlanMaterialTransactionSummary struct {
	JobProductionPlanID string                                            `json:"job_production_plan_id,omitempty"`
	TonsByMaterialType  []jobProductionPlanMaterialTransactionSummaryItem `json:"tons_by_material_type"`
}

func newDoJobProductionPlanMaterialTransactionSummariesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan material transaction summary",
		Long: `Create a job production plan material transaction summary.

Returns accepted material transaction tons grouped by material type for a single
job production plan. Only accepted material transactions measured in tons are
included in the summary.

Required flags:
  --job-production-plan  Job production plan ID (required)`,
		Example: `  # Summarize material transactions for a job production plan
  xbe do job-production-plan-material-transaction-summaries create --job-production-plan 123

  # JSON output
  xbe do job-production-plan-material-transaction-summaries create --job-production-plan 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanMaterialTransactionSummariesCreate,
	}
	initDoJobProductionPlanMaterialTransactionSummariesCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanMaterialTransactionSummariesCmd.AddCommand(newDoJobProductionPlanMaterialTransactionSummariesCreateCmd())
}

func initDoJobProductionPlanMaterialTransactionSummariesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanMaterialTransactionSummariesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanMaterialTransactionSummariesCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	if strings.TrimSpace(opts.JobProductionPlan) == "" {
		err := fmt.Errorf("--job-production-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-material-transaction-summaries",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-material-transaction-summaries", jsonBody)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	summary := jobProductionPlanMaterialTransactionSummaryFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), summary)
	}

	return renderJobProductionPlanMaterialTransactionSummary(cmd, summary)
}

func parseDoJobProductionPlanMaterialTransactionSummariesCreateOptions(cmd *cobra.Command) (doJobProductionPlanMaterialTransactionSummariesCreateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doJobProductionPlanMaterialTransactionSummariesCreateOptions{}, err
	}
	jobProductionPlan, err := cmd.Flags().GetString("job-production-plan")
	if err != nil {
		return doJobProductionPlanMaterialTransactionSummariesCreateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doJobProductionPlanMaterialTransactionSummariesCreateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doJobProductionPlanMaterialTransactionSummariesCreateOptions{}, err
	}

	return doJobProductionPlanMaterialTransactionSummariesCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		JobProductionPlan: jobProductionPlan,
	}, nil
}

func jobProductionPlanMaterialTransactionSummaryFromSingle(resp jsonAPISingleResponse) jobProductionPlanMaterialTransactionSummary {
	attrs := resp.Data.Attributes
	summary := jobProductionPlanMaterialTransactionSummary{}

	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		summary.JobProductionPlanID = rel.Data.ID
	}

	summary.JobProductionPlanID = firstNonEmpty(
		summary.JobProductionPlanID,
		stringAttr(attrs, "job-production-plan-id"),
		stringAttr(attrs, "job_production_plan_id"),
	)

	summary.TonsByMaterialType = parseJobProductionPlanMaterialTransactionSummaryItems(attrs)
	return summary
}

func parseJobProductionPlanMaterialTransactionSummaryItems(attrs map[string]any) []jobProductionPlanMaterialTransactionSummaryItem {
	if attrs == nil {
		return []jobProductionPlanMaterialTransactionSummaryItem{}
	}

	raw, ok := attrs["tons-by-material-type"]
	if !ok {
		raw = attrs["tons_by_material_type"]
	}
	items, ok := raw.([]any)
	if !ok || len(items) == 0 {
		return []jobProductionPlanMaterialTransactionSummaryItem{}
	}

	rows := make([]jobProductionPlanMaterialTransactionSummaryItem, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}

		row := jobProductionPlanMaterialTransactionSummaryItem{
			MaterialTypeID: firstNonEmpty(
				stringAttr(entry, "material_type_id"),
				stringAttr(entry, "material-type-id"),
			),
			MaterialTypeFullyQualifiedName: firstNonEmpty(
				stringAttr(entry, "material_type_fully_qualified_name"),
				stringAttr(entry, "material-type-fully-qualified-name"),
			),
			Tons: formatAnyValue(entry["tons"]),
		}

		if row.Tons == "" {
			row.Tons = formatAnyValue(entry["tons_sum"])
		}

		rows = append(rows, row)
	}

	return rows
}

func renderJobProductionPlanMaterialTransactionSummary(cmd *cobra.Command, summary jobProductionPlanMaterialTransactionSummary) error {
	out := cmd.OutOrStdout()

	fmt.Fprintln(out, "Job Production Plan Material Transaction Summary")
	if summary.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan: %s\n", summary.JobProductionPlanID)
	}

	if len(summary.TonsByMaterialType) == 0 {
		fmt.Fprintln(out, "No material transactions found for this job production plan.")
		return nil
	}

	fmt.Fprintln(out)
	writer := tabwriter.NewWriter(out, 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "MATERIAL TYPE\tTONS")
	for _, row := range summary.TonsByMaterialType {
		name := strings.TrimSpace(row.MaterialTypeFullyQualifiedName)
		if name == "" {
			name = strings.TrimSpace(row.MaterialTypeID)
		}
		fmt.Fprintf(writer, "%s\t%s\n", name, strings.TrimSpace(row.Tons))
	}
	return writer.Flush()
}
