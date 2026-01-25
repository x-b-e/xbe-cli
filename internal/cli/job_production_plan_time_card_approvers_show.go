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

type jobProductionPlanTimeCardApproversShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanTimeCardApproverDetails struct {
	ID                  string `json:"id"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
	JobProductionPlan   string `json:"job_production_plan,omitempty"`
	UserID              string `json:"user_id,omitempty"`
	UserName            string `json:"user_name,omitempty"`
	UserEmail           string `json:"user_email,omitempty"`
	UserMobile          string `json:"user_mobile,omitempty"`
}

func newJobProductionPlanTimeCardApproversShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan time card approver details",
		Long: `Show the full details of a job production plan time card approver.

Output Fields:
  ID                  Time card approver identifier
  Job Production Plan Job production plan
  User                Approver user
  User Email          Approver email address
  User Mobile         Approver mobile number

Arguments:
  <id>    Job production plan time card approver ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a time card approver
  xbe view job-production-plan-time-card-approvers show 123

  # JSON output
  xbe view job-production-plan-time-card-approvers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanTimeCardApproversShow,
	}
	initJobProductionPlanTimeCardApproversShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanTimeCardApproversCmd.AddCommand(newJobProductionPlanTimeCardApproversShowCmd())
}

func initJobProductionPlanTimeCardApproversShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanTimeCardApproversShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanTimeCardApproversShowOptions(cmd)
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
		return fmt.Errorf("job production plan time card approver id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-time-card-approvers]", "job-production-plan,user")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[users]", "name,email-address,mobile-number")
	query.Set("include", "job-production-plan,user")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-time-card-approvers/"+id, query)
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

	details := buildJobProductionPlanTimeCardApproverDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanTimeCardApproverDetails(cmd, details)
}

func parseJobProductionPlanTimeCardApproversShowOptions(cmd *cobra.Command) (jobProductionPlanTimeCardApproversShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanTimeCardApproversShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanTimeCardApproverDetails(resp jsonAPISingleResponse) jobProductionPlanTimeCardApproverDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := jobProductionPlanTimeCardApproverDetails{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
		if plan, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			jobNumber := stringAttr(plan.Attributes, "job-number")
			jobName := stringAttr(plan.Attributes, "job-name")
			if jobNumber != "" && jobName != "" {
				details.JobProductionPlan = fmt.Sprintf("%s - %s", jobNumber, jobName)
			} else {
				details.JobProductionPlan = firstNonEmpty(jobNumber, jobName)
			}
		}
	}

	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UserName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			details.UserEmail = strings.TrimSpace(stringAttr(user.Attributes, "email-address"))
			details.UserMobile = strings.TrimSpace(stringAttr(user.Attributes, "mobile-number"))
		}
	}

	return details
}

func renderJobProductionPlanTimeCardApproverDetails(cmd *cobra.Command, details jobProductionPlanTimeCardApproverDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	writeLabelWithID(out, "Job Production Plan", details.JobProductionPlan, details.JobProductionPlanID)

	userDisplay := firstNonEmpty(details.UserName, details.UserEmail)
	writeLabelWithID(out, "User", userDisplay, details.UserID)

	if details.UserEmail != "" {
		fmt.Fprintf(out, "User Email: %s\n", details.UserEmail)
	}
	if details.UserMobile != "" {
		fmt.Fprintf(out, "User Mobile: %s\n", details.UserMobile)
	}

	return nil
}
