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

type jobProductionPlanRecapsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanRecapDetails struct {
	ID                    string `json:"id"`
	PlanID                string `json:"plan_id,omitempty"`
	Plan                  string `json:"plan,omitempty"`
	Markdown              string `json:"markdown,omitempty"`
	RelatedPastMarkdown   string `json:"related_past_markdown,omitempty"`
	RelatedFutureMarkdown string `json:"related_future_markdown,omitempty"`
	CreatedAt             string `json:"created_at,omitempty"`
	UpdatedAt             string `json:"updated_at,omitempty"`
}

func newJobProductionPlanRecapsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan recap details",
		Long: `Show the full details of a job production plan recap.

Output Fields:
  ID               Recap identifier
  Plan             Job production plan name/number
  Created          Created timestamp
  Updated          Updated timestamp
  Recap            Recap markdown
  Related Past     Related past recap markdown
  Related Future   Related future recap markdown

Arguments:
  <id>  The job production plan recap ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show recap details
  xbe view job-production-plan-recaps show 123

  # Show as JSON
  xbe view job-production-plan-recaps show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanRecapsShow,
	}
	initJobProductionPlanRecapsShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanRecapsCmd.AddCommand(newJobProductionPlanRecapsShowCmd())
}

func initJobProductionPlanRecapsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanRecapsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanRecapsShowOptions(cmd)
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
		return fmt.Errorf("job production plan recap id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-recaps]", "created-at,updated-at,plan,markdown,related-past-markdown,related-future-markdown")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("include", "plan")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-recaps/"+id, query)
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

	details := buildJobProductionPlanRecapDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanRecapDetails(cmd, details)
}

func parseJobProductionPlanRecapsShowOptions(cmd *cobra.Command) (jobProductionPlanRecapsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanRecapsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanRecapDetails(resp jsonAPISingleResponse) jobProductionPlanRecapDetails {
	attrs := resp.Data.Attributes
	details := jobProductionPlanRecapDetails{
		ID:                    resp.Data.ID,
		Markdown:              strings.TrimSpace(stringAttr(attrs, "markdown")),
		RelatedPastMarkdown:   strings.TrimSpace(stringAttr(attrs, "related-past-markdown")),
		RelatedFutureMarkdown: strings.TrimSpace(stringAttr(attrs, "related-future-markdown")),
		CreatedAt:             formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:             formatDateTime(stringAttr(attrs, "updated-at")),
	}

	included := make(map[string]jsonAPIResource)
	for _, resource := range resp.Included {
		included[resourceKey(resource.Type, resource.ID)] = resource
	}

	planType := ""
	if rel, ok := resp.Data.Relationships["plan"]; ok && rel.Data != nil {
		details.PlanID = rel.Data.ID
		planType = rel.Data.Type
	}

	if details.PlanID != "" && planType != "" {
		if plan, ok := included[resourceKey(planType, details.PlanID)]; ok {
			jobNumber := strings.TrimSpace(stringAttr(plan.Attributes, "job-number"))
			jobName := strings.TrimSpace(stringAttr(plan.Attributes, "job-name"))
			if jobNumber != "" && jobName != "" {
				details.Plan = fmt.Sprintf("%s - %s", jobNumber, jobName)
			} else {
				details.Plan = firstNonEmpty(jobNumber, jobName)
			}
		}
	}

	return details
}

func renderJobProductionPlanRecapDetails(cmd *cobra.Command, details jobProductionPlanRecapDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Plan != "" {
		fmt.Fprintf(out, "Plan: %s\n", details.Plan)
	} else if details.PlanID != "" {
		fmt.Fprintf(out, "Plan ID: %s\n", details.PlanID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated: %s\n", details.UpdatedAt)
	}

	if details.Markdown != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Recap:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, cleanMarkdownFull(details.Markdown))
	}
	if details.RelatedPastMarkdown != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Related Past:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, cleanMarkdownFull(details.RelatedPastMarkdown))
	}
	if details.RelatedFutureMarkdown != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Related Future:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, cleanMarkdownFull(details.RelatedFutureMarkdown))
	}

	return nil
}
