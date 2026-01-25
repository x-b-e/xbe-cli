package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doJobProductionPlanSafetyRisksSuggestionsCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	JobProductionPlanID string
	Options             string
	IsAsync             bool
}

func newDoJobProductionPlanSafetyRisksSuggestionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan safety risks suggestion",
		Long: `Create a job production plan safety risks suggestion.

Required:
  --job-production-plan  Job production plan ID

Optional:
  --options   JSON object with generation options
  --is-async  Generate asynchronously (default true)

Notes:
  Setting --is-async=false will block until risks are generated.`,
		Example: `  # Generate safety risks suggestions
  xbe do job-production-plan-safety-risks-suggestions create --job-production-plan 123

  # Provide options
  xbe do job-production-plan-safety-risks-suggestions create \
    --job-production-plan 123 \
    --options '{"include_other_incidents":true}'

  # Generate synchronously
  xbe do job-production-plan-safety-risks-suggestions create \
    --job-production-plan 123 \
    --is-async=false`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanSafetyRisksSuggestionsCreate,
	}
	initDoJobProductionPlanSafetyRisksSuggestionsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanSafetyRisksSuggestionsCmd.AddCommand(newDoJobProductionPlanSafetyRisksSuggestionsCreateCmd())
}

func initDoJobProductionPlanSafetyRisksSuggestionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("options", "", "Generation options JSON")
	cmd.Flags().Bool("is-async", true, "Generate asynchronously")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanSafetyRisksSuggestionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanSafetyRisksSuggestionsCreateOptions(cmd)
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

	if opts.JobProductionPlanID == "" {
		err := fmt.Errorf("--job-production-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"is-async": opts.IsAsync,
	}

	if strings.TrimSpace(opts.Options) != "" {
		options, err := parseOptionsJSON(opts.Options)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["options"] = options
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlanID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-safety-risks-suggestions",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-safety-risks-suggestions", jsonBody)
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

	row := buildJobProductionPlanSafetyRisksSuggestionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan safety risks suggestion %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanSafetyRisksSuggestionsCreateOptions(cmd *cobra.Command) (doJobProductionPlanSafetyRisksSuggestionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	options, _ := cmd.Flags().GetString("options")
	isAsync, _ := cmd.Flags().GetBool("is-async")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanSafetyRisksSuggestionsCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		JobProductionPlanID: jobProductionPlanID,
		Options:             options,
		IsAsync:             isAsync,
	}, nil
}

func parseOptionsJSON(raw string) (map[string]any, error) {
	var decoded any
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return nil, fmt.Errorf("--options must be valid JSON: %w", err)
	}
	options, ok := decoded.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("--options must be a JSON object")
	}
	return options, nil
}

func buildJobProductionPlanSafetyRisksSuggestionRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanSafetyRisksSuggestionRow {
	resource := resp.Data
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
	return row
}
