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

type jobProductionPlanSafetyRisksSuggestionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanSafetyRisksSuggestionDetails struct {
	ID                  string         `json:"id"`
	JobProductionPlanID string         `json:"job_production_plan_id,omitempty"`
	Options             map[string]any `json:"options,omitempty"`
	IsAsync             bool           `json:"is_async"`
	Prompt              string         `json:"prompt,omitempty"`
	Response            string         `json:"response,omitempty"`
	IsFulfilled         bool           `json:"is_fulfilled"`
	Risks               []string       `json:"risks,omitempty"`
}

func newJobProductionPlanSafetyRisksSuggestionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan safety risks suggestion details",
		Long: `Show the full details of a job production plan safety risks suggestion.

Output Fields:
  ID                  Suggestion identifier
  Job Production Plan Associated job production plan ID
  Options             Generation options payload
  Is Async            Whether suggestions were generated asynchronously
  Is Fulfilled        Whether suggestions have been generated
  Risks               Generated safety risks
  Prompt              Prompt sent for generation
  Response            Raw model response

Arguments:
  <id>    The suggestion ID (required). You can find IDs using the list command.`,
		Example: `  # Show a safety risks suggestion
  xbe view job-production-plan-safety-risks-suggestions show 123

  # Get JSON output
  xbe view job-production-plan-safety-risks-suggestions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanSafetyRisksSuggestionsShow,
	}
	initJobProductionPlanSafetyRisksSuggestionsShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanSafetyRisksSuggestionsCmd.AddCommand(newJobProductionPlanSafetyRisksSuggestionsShowCmd())
}

func initJobProductionPlanSafetyRisksSuggestionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanSafetyRisksSuggestionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanSafetyRisksSuggestionsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("job production plan safety risks suggestion id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-safety-risks-suggestions/"+id, nil)
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

	details := buildJobProductionPlanSafetyRisksSuggestionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanSafetyRisksSuggestionDetails(cmd, details)
}

func parseJobProductionPlanSafetyRisksSuggestionsShowOptions(cmd *cobra.Command) (jobProductionPlanSafetyRisksSuggestionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanSafetyRisksSuggestionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanSafetyRisksSuggestionDetails(resp jsonAPISingleResponse) jobProductionPlanSafetyRisksSuggestionDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := jobProductionPlanSafetyRisksSuggestionDetails{
		ID:          resource.ID,
		Options:     mapAttr(attrs, "options"),
		IsAsync:     boolAttr(attrs, "is-async"),
		Prompt:      stringAttr(attrs, "prompt"),
		Response:    stringAttr(attrs, "response"),
		IsFulfilled: boolAttr(attrs, "is-fulfilled"),
		Risks:       stringSliceAttr(attrs, "risks"),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
	}

	return details
}

func renderJobProductionPlanSafetyRisksSuggestionDetails(cmd *cobra.Command, details jobProductionPlanSafetyRisksSuggestionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan: %s\n", details.JobProductionPlanID)
	}
	fmt.Fprintf(out, "Is Async: %t\n", details.IsAsync)
	fmt.Fprintf(out, "Is Fulfilled: %t\n", details.IsFulfilled)

	if len(details.Options) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Options:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, formatJSON(details.Options))
	}

	if len(details.Risks) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Risks:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, risk := range details.Risks {
			if strings.TrimSpace(risk) == "" {
				continue
			}
			fmt.Fprintf(out, "- %s\n", risk)
		}
	}

	if strings.TrimSpace(details.Prompt) != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Prompt:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Prompt)
	}

	if strings.TrimSpace(details.Response) != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Response:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Response)
	}

	return nil
}

func formatJSON(value any) string {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(encoded)
}
