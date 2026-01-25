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

type jobProductionPlanSubscriptionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanSubscriptionDetails struct {
	ID                              string `json:"id"`
	JobProductionPlanID             string `json:"job_production_plan_id,omitempty"`
	JobProductionPlanJobNumber      string `json:"job_number,omitempty"`
	JobProductionPlanJobName        string `json:"job_name,omitempty"`
	UserID                          string `json:"user_id,omitempty"`
	UserName                        string `json:"user_name,omitempty"`
	UserEmail                       string `json:"user_email,omitempty"`
	ContactMethod                   string `json:"contact_method,omitempty"`
	CalculatedContactMethod         string `json:"calculated_contact_method,omitempty"`
	IsCreatedViaProjectSubscription bool   `json:"is_created_via_project_subscription"`
	CreatedViaType                  string `json:"created_via_type,omitempty"`
	CreatedViaID                    string `json:"created_via_id,omitempty"`
	ProjectSubscriptionID           string `json:"project_subscription_id,omitempty"`
}

func newJobProductionPlanSubscriptionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan subscription details",
		Long: `Show the full details of a job production plan subscription.

Includes the associated job production plan and user information.

Arguments:
  <id>  The job production plan subscription ID (required).`,
		Example: `  # Show a subscription
  xbe view job-production-plan-subscriptions show 123

  # Output as JSON
  xbe view job-production-plan-subscriptions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanSubscriptionsShow,
	}
	initJobProductionPlanSubscriptionsShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanSubscriptionsCmd.AddCommand(newJobProductionPlanSubscriptionsShowCmd())
}

func initJobProductionPlanSubscriptionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanSubscriptionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanSubscriptionsShowOptions(cmd)
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
		return fmt.Errorf("job production plan subscription id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-subscriptions]", "job-production-plan,user,contact-method,calculated-contact-method,is-created-via-project-subscription,created-via,project-subscription")
	query.Set("include", "job-production-plan,user")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[users]", "name,email-address")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-subscriptions/"+id, query)
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

	details := buildJobProductionPlanSubscriptionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanSubscriptionDetails(cmd, details)
}

func parseJobProductionPlanSubscriptionsShowOptions(cmd *cobra.Command) (jobProductionPlanSubscriptionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanSubscriptionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanSubscriptionDetails(resp jsonAPISingleResponse) jobProductionPlanSubscriptionDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := jobProductionPlanSubscriptionDetails{
		ID: resp.Data.ID,
	}

	details.ContactMethod = stringAttr(resp.Data.Attributes, "contact-method")
	details.CalculatedContactMethod = stringAttr(resp.Data.Attributes, "calculated-contact-method")
	details.IsCreatedViaProjectSubscription = boolAttr(resp.Data.Attributes, "is-created-via-project-subscription")

	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
		if plan, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.JobProductionPlanJobNumber = stringAttr(plan.Attributes, "job-number")
			details.JobProductionPlanJobName = stringAttr(plan.Attributes, "job-name")
		}
	}

	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UserName = stringAttr(user.Attributes, "name")
			details.UserEmail = stringAttr(user.Attributes, "email-address")
		}
	}

	if rel, ok := resp.Data.Relationships["created-via"]; ok && rel.Data != nil {
		details.CreatedViaType = rel.Data.Type
		details.CreatedViaID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["project-subscription"]; ok && rel.Data != nil {
		details.ProjectSubscriptionID = rel.Data.ID
	}

	return details
}

func renderJobProductionPlanSubscriptionDetails(cmd *cobra.Command, details jobProductionPlanSubscriptionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan: %s\n", details.JobProductionPlanID)
	}
	if details.JobProductionPlanJobNumber != "" {
		fmt.Fprintf(out, "Job Number: %s\n", details.JobProductionPlanJobNumber)
	}
	if details.JobProductionPlanJobName != "" {
		fmt.Fprintf(out, "Job Name: %s\n", details.JobProductionPlanJobName)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	if details.UserName != "" {
		fmt.Fprintf(out, "User Name: %s\n", details.UserName)
	}
	if details.UserEmail != "" {
		fmt.Fprintf(out, "User Email: %s\n", details.UserEmail)
	}
	if details.ContactMethod != "" {
		fmt.Fprintf(out, "Contact Method: %s\n", details.ContactMethod)
	}
	if details.CalculatedContactMethod != "" {
		fmt.Fprintf(out, "Calculated Contact Method: %s\n", details.CalculatedContactMethod)
	}
	fmt.Fprintf(out, "Is Created Via Project Subscription: %t\n", details.IsCreatedViaProjectSubscription)
	if details.CreatedViaType != "" || details.CreatedViaID != "" {
		fmt.Fprintf(out, "Created Via: %s %s\n", details.CreatedViaType, details.CreatedViaID)
	}
	if details.ProjectSubscriptionID != "" {
		fmt.Fprintf(out, "Project Subscription: %s\n", details.ProjectSubscriptionID)
	}

	return nil
}
