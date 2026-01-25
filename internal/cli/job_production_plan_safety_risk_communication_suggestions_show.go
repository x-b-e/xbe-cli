package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type jobProductionPlanSafetyRiskCommunicationSuggestionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanSafetyRiskCommunicationSuggestionDetails struct {
	ID                  string `json:"id"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
	JobProductionPlan   string `json:"job_production_plan,omitempty"`
	IsAsync             bool   `json:"is_async"`
	IsFulfilled         bool   `json:"is_fulfilled"`
	Prompt              string `json:"prompt,omitempty"`
	Response            string `json:"response,omitempty"`
	Suggestion          string `json:"suggestion,omitempty"`
	Options             any    `json:"options,omitempty"`
	CreatedAt           string `json:"created_at,omitempty"`
	UpdatedAt           string `json:"updated_at,omitempty"`
}

func newJobProductionPlanSafetyRiskCommunicationSuggestionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show safety risk communication suggestion details",
		Long: `Show the full details of a job production plan safety risk communication suggestion.

Output Fields:
  ID                   Suggestion identifier
  Job Production Plan  Job production plan name/number
  Async                Whether the suggestion was generated asynchronously
  Fulfilled            Whether the suggestion has completed generation
  Created              Created timestamp
  Updated              Updated timestamp
  Options              Options used for generation
  Prompt               Prompt sent to the generator
  Suggestion           Generated communication plan
  Response             Raw response payload

Arguments:
  <id>  The suggestion ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show suggestion details
  xbe view job-production-plan-safety-risk-communication-suggestions show 123

  # Show as JSON
  xbe view job-production-plan-safety-risk-communication-suggestions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanSafetyRiskCommunicationSuggestionsShow,
	}
	initJobProductionPlanSafetyRiskCommunicationSuggestionsShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanSafetyRiskCommunicationSuggestionsCmd.AddCommand(newJobProductionPlanSafetyRiskCommunicationSuggestionsShowCmd())
}

func initJobProductionPlanSafetyRiskCommunicationSuggestionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanSafetyRiskCommunicationSuggestionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanSafetyRiskCommunicationSuggestionsShowOptions(cmd)
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
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("job production plan safety risk communication suggestion id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-safety-risk-communication-suggestions]", "created-at,updated-at,job-production-plan,is-async,is-fulfilled,options,prompt,response,suggestion")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("include", "job-production-plan")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-safety-risk-communication-suggestions/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildJobProductionPlanSafetyRiskCommunicationSuggestionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanSafetyRiskCommunicationSuggestionDetails(cmd, details)
}

func parseJobProductionPlanSafetyRiskCommunicationSuggestionsShowOptions(cmd *cobra.Command) (jobProductionPlanSafetyRiskCommunicationSuggestionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanSafetyRiskCommunicationSuggestionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanSafetyRiskCommunicationSuggestionDetails(resp jsonAPISingleResponse) jobProductionPlanSafetyRiskCommunicationSuggestionDetails {
	attrs := resp.Data.Attributes
	details := jobProductionPlanSafetyRiskCommunicationSuggestionDetails{
		ID:          resp.Data.ID,
		IsAsync:     boolAttr(attrs, "is-async"),
		IsFulfilled: boolAttr(attrs, "is-fulfilled"),
		Prompt:      strings.TrimSpace(stringAttr(attrs, "prompt")),
		Response:    strings.TrimSpace(stringAttr(attrs, "response")),
		Suggestion:  strings.TrimSpace(stringAttr(attrs, "suggestion")),
		Options:     attrs["options"],
		CreatedAt:   formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:   formatDateTime(stringAttr(attrs, "updated-at")),
	}

	jppType := ""
	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
		jppType = rel.Data.Type
	}

	if len(resp.Included) == 0 {
		return details
	}

	included := make(map[string]jsonAPIResource)
	for _, resource := range resp.Included {
		included[resourceKey(resource.Type, resource.ID)] = resource
	}

	if details.JobProductionPlanID != "" && jppType != "" {
		if jpp, ok := included[resourceKey(jppType, details.JobProductionPlanID)]; ok {
			jobNumber := strings.TrimSpace(stringAttr(jpp.Attributes, "job-number"))
			jobName := strings.TrimSpace(stringAttr(jpp.Attributes, "job-name"))
			if jobNumber != "" && jobName != "" {
				details.JobProductionPlan = fmt.Sprintf("%s - %s", jobNumber, jobName)
			} else {
				details.JobProductionPlan = firstNonEmpty(jobNumber, jobName)
			}
		}
	}

	return details
}

func renderJobProductionPlanSafetyRiskCommunicationSuggestionDetails(cmd *cobra.Command, details jobProductionPlanSafetyRiskCommunicationSuggestionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlan != "" {
		fmt.Fprintf(out, "Job Production Plan: %s\n", details.JobProductionPlan)
	} else if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlanID)
	}
	fmt.Fprintf(out, "Async: %s\n", formatSuggestionBool(details.IsAsync))
	fmt.Fprintf(out, "Fulfilled: %s\n", formatSuggestionBool(details.IsFulfilled))
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated: %s\n", details.UpdatedAt)
	}
	if details.Options != nil {
		fmt.Fprintln(out, "Options:")
		fmt.Fprintln(out, formatJobProductionPlanSafetyRiskCommunicationSuggestionJSON(details.Options))
	}

	if details.Prompt != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Prompt:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Prompt)
	}
	if details.Suggestion != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Suggestion:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Suggestion)
	}
	if details.Response != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Response:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Response)
	}

	return nil
}

func formatJobProductionPlanSafetyRiskCommunicationSuggestionJSON(value any) string {
	if value == nil {
		return ""
	}
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(pretty)
}

func formatSuggestionBool(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
