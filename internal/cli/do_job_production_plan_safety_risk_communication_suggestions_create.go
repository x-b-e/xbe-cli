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

type doJobProductionPlanSafetyRiskCommunicationSuggestionsCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	JobProductionPlan string
	IsAsync           bool
	Options           string
}

func newDoJobProductionPlanSafetyRiskCommunicationSuggestionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Generate a safety risk communication suggestion",
		Long: `Generate a safety risk communication suggestion for a job production plan.

Required flags:
  --job-production-plan  Job production plan ID (required)

Optional flags:
  --is-async              Generate asynchronously (default true)
  --options               Options JSON object passed to the generator

Notes:
  When --is-async is true (default), the suggestion is generated asynchronously.
  Use the show command to check the fulfilled status.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create an async suggestion (default)
  xbe do job-production-plan-safety-risk-communication-suggestions create \
    --job-production-plan 123

  # Create with options
  xbe do job-production-plan-safety-risk-communication-suggestions create \
    --job-production-plan 123 \
    --options '{"temperature":0.2}'

  # Create synchronously
  xbe do job-production-plan-safety-risk-communication-suggestions create \
    --job-production-plan 123 \
    --is-async=false`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanSafetyRiskCommunicationSuggestionsCreate,
	}
	initDoJobProductionPlanSafetyRiskCommunicationSuggestionsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanSafetyRiskCommunicationSuggestionsCmd.AddCommand(newDoJobProductionPlanSafetyRiskCommunicationSuggestionsCreateCmd())
}

func initDoJobProductionPlanSafetyRiskCommunicationSuggestionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().Bool("is-async", true, "Generate asynchronously")
	cmd.Flags().String("options", "", "Options JSON object passed to the generator")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanSafetyRiskCommunicationSuggestionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanSafetyRiskCommunicationSuggestionsCreateOptions(cmd)
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

	if opts.JobProductionPlan == "" {
		err := fmt.Errorf("--job-production-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"is-async": opts.IsAsync,
	}

	if opts.Options != "" {
		var parsed any
		if err := json.Unmarshal([]byte(opts.Options), &parsed); err != nil {
			return fmt.Errorf("invalid --options JSON: %w", err)
		}
		attributes["options"] = parsed
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
			"type":          "job-production-plan-safety-risk-communication-suggestions",
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-safety-risk-communication-suggestions", jsonBody)
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

	row := buildJobProductionPlanSafetyRiskCommunicationSuggestionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan safety risk communication suggestion %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanSafetyRiskCommunicationSuggestionsCreateOptions(cmd *cobra.Command) (doJobProductionPlanSafetyRiskCommunicationSuggestionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	isAsync, _ := cmd.Flags().GetBool("is-async")
	options, _ := cmd.Flags().GetString("options")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanSafetyRiskCommunicationSuggestionsCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		JobProductionPlan: jobProductionPlan,
		IsAsync:           isAsync,
		Options:           options,
	}, nil
}

func buildJobProductionPlanSafetyRiskCommunicationSuggestionRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanSafetyRiskCommunicationSuggestionRow {
	included := make(map[string]jsonAPIResource)
	for _, resource := range resp.Included {
		included[resourceKey(resource.Type, resource.ID)] = resource
	}

	return buildJobProductionPlanSafetyRiskCommunicationSuggestionRow(resp.Data, included)
}
