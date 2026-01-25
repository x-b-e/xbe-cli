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

type jobProductionPlanBroadcastMessagesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanBroadcastMessageDetails struct {
	ID                  string   `json:"id"`
	JobProductionPlan   string   `json:"job_production_plan,omitempty"`
	JobProductionPlanID string   `json:"job_production_plan_id,omitempty"`
	CreatedBy           string   `json:"created_by,omitempty"`
	CreatedByID         string   `json:"created_by_id,omitempty"`
	Message             string   `json:"message,omitempty"`
	Summary             string   `json:"summary,omitempty"`
	UserIDs             []string `json:"user_ids,omitempty"`
	IsHidden            bool     `json:"is_hidden"`
	CreatedAt           string   `json:"created_at,omitempty"`
	UpdatedAt           string   `json:"updated_at,omitempty"`
}

func newJobProductionPlanBroadcastMessagesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan broadcast message details",
		Long: `Show the full details of a job production plan broadcast message.

Output Fields:
  ID                     Broadcast message identifier
  Job Production Plan    Job production plan name/number
  Created By             User who created the message
  Hidden                 Whether the message is hidden
  Summary                Message summary
  Message                Full message content
  User IDs               Recipient user IDs (if specified)
  Created                Created timestamp
  Updated                Updated timestamp

Arguments:
  <id>  The broadcast message ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show message details
  xbe view job-production-plan-broadcast-messages show 123

  # Show as JSON
  xbe view job-production-plan-broadcast-messages show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanBroadcastMessagesShow,
	}
	initJobProductionPlanBroadcastMessagesShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanBroadcastMessagesCmd.AddCommand(newJobProductionPlanBroadcastMessagesShowCmd())
}

func initJobProductionPlanBroadcastMessagesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanBroadcastMessagesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanBroadcastMessagesShowOptions(cmd)
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
		return fmt.Errorf("broadcast message id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-broadcast-messages]", "message,summary,user-ids,is-hidden,created-at,updated-at,job-production-plan,created-by")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[users]", "name")
	query.Set("include", "job-production-plan,created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-broadcast-messages/"+id, query)
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

	details := buildJobProductionPlanBroadcastMessageDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanBroadcastMessageDetails(cmd, details)
}

func parseJobProductionPlanBroadcastMessagesShowOptions(cmd *cobra.Command) (jobProductionPlanBroadcastMessagesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanBroadcastMessagesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanBroadcastMessageDetails(resp jsonAPISingleResponse) jobProductionPlanBroadcastMessageDetails {
	attrs := resp.Data.Attributes
	details := jobProductionPlanBroadcastMessageDetails{
		ID:        resp.Data.ID,
		Message:   strings.TrimSpace(stringAttr(attrs, "message")),
		Summary:   strings.TrimSpace(stringAttr(attrs, "summary")),
		UserIDs:   stringSliceAttr(attrs, "user-ids"),
		IsHidden:  boolAttr(attrs, "is-hidden"),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	jppType := ""
	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
		jppType = rel.Data.Type
	}
	createdByType := ""
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		createdByType = rel.Data.Type
	}

	if len(resp.Included) == 0 {
		return details
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
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

	if details.CreatedByID != "" && createdByType != "" {
		if user, ok := included[resourceKey(createdByType, details.CreatedByID)]; ok {
			details.CreatedBy = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	return details
}

func renderJobProductionPlanBroadcastMessageDetails(cmd *cobra.Command, details jobProductionPlanBroadcastMessageDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlan != "" {
		if details.JobProductionPlanID != "" {
			fmt.Fprintf(out, "Job Production Plan: %s (%s)\n", details.JobProductionPlan, details.JobProductionPlanID)
		} else {
			fmt.Fprintf(out, "Job Production Plan: %s\n", details.JobProductionPlan)
		}
	} else if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlanID)
	}
	if details.CreatedBy != "" {
		if details.CreatedByID != "" {
			fmt.Fprintf(out, "Created By: %s (%s)\n", details.CreatedBy, details.CreatedByID)
		} else {
			fmt.Fprintf(out, "Created By: %s\n", details.CreatedBy)
		}
	} else if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By ID: %s\n", details.CreatedByID)
	}
	fmt.Fprintf(out, "Hidden: %t\n", details.IsHidden)
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated: %s\n", details.UpdatedAt)
	}

	if details.Summary != "" {
		fmt.Fprintf(out, "Summary: %s\n", details.Summary)
	}
	if len(details.UserIDs) > 0 {
		fmt.Fprintf(out, "User IDs: %s\n", strings.Join(details.UserIDs, ", "))
	}

	if details.Message != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Message:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Message)
	}

	return nil
}
