package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type jobProductionPlanStatusChangesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanStatusChangeDetails struct {
	ID                                        string `json:"id"`
	JobProductionPlan                         string `json:"job_production_plan,omitempty"`
	JobProductionPlanID                       string `json:"job_production_plan_id,omitempty"`
	Status                                    string `json:"status,omitempty"`
	Comment                                   string `json:"comment,omitempty"`
	ChangedAt                                 string `json:"changed_at,omitempty"`
	ChangedBy                                 string `json:"changed_by,omitempty"`
	ChangedByID                               string `json:"changed_by_id,omitempty"`
	JobProductionPlanCancellationReasonType   string `json:"job_production_plan_cancellation_reason_type,omitempty"`
	JobProductionPlanCancellationReasonTypeID string `json:"job_production_plan_cancellation_reason_type_id,omitempty"`
	CreatedAt                                 string `json:"created_at,omitempty"`
	UpdatedAt                                 string `json:"updated_at,omitempty"`
}

func newJobProductionPlanStatusChangesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan status change details",
		Long: `Show the full details of a job production plan status change.

Output Fields:
  ID                     Status change identifier
  Job Production Plan    Job production plan name/number
  Status                 New status
  Changed By             User who changed the status
  Changed At             When the status changed
  Comment                Status change comment (if provided)
  Cancellation Reason    Cancellation reason type (if applicable)
  Created                Created timestamp
  Updated                Updated timestamp

Arguments:
  <id>  The status change ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show status change details
  xbe view job-production-plan-status-changes show 123

  # Show as JSON
  xbe view job-production-plan-status-changes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanStatusChangesShow,
	}
	initJobProductionPlanStatusChangesShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanStatusChangesCmd.AddCommand(newJobProductionPlanStatusChangesShowCmd())
}

func initJobProductionPlanStatusChangesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanStatusChangesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseJobProductionPlanStatusChangesShowOptions(cmd)
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
		return fmt.Errorf("status change id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-status-changes]", "status,comment,changed-at,created-at,updated-at,job-production-plan,changed-by,job-production-plan-cancellation-reason-type")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[users]", "name")
	query.Set("fields[job-production-plan-cancellation-reason-types]", "name,description")
	query.Set("include", "job-production-plan,changed-by,job-production-plan-cancellation-reason-type")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-status-changes/"+id, query)
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

	details := buildJobProductionPlanStatusChangeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanStatusChangeDetails(cmd, details)
}

func parseJobProductionPlanStatusChangesShowOptions(cmd *cobra.Command) (jobProductionPlanStatusChangesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanStatusChangesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanStatusChangeDetails(resp jsonAPISingleResponse) jobProductionPlanStatusChangeDetails {
	included := make(map[string]jsonAPIResource)
	for _, resource := range resp.Included {
		included[resourceKey(resource.Type, resource.ID)] = resource
	}

	details := jobProductionPlanStatusChangeDetails{
		ID:        resp.Data.ID,
		Status:    strings.TrimSpace(stringAttr(resp.Data.Attributes, "status")),
		Comment:   strings.TrimSpace(stringAttr(resp.Data.Attributes, "comment")),
		ChangedAt: formatDateTime(stringAttr(resp.Data.Attributes, "changed-at")),
		CreatedAt: formatDateTime(stringAttr(resp.Data.Attributes, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(resp.Data.Attributes, "updated-at")),
	}

	jppType := ""
	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
		jppType = rel.Data.Type
	}
	changedByType := ""
	if rel, ok := resp.Data.Relationships["changed-by"]; ok && rel.Data != nil {
		details.ChangedByID = rel.Data.ID
		changedByType = rel.Data.Type
	}
	reasonType := ""
	if rel, ok := resp.Data.Relationships["job-production-plan-cancellation-reason-type"]; ok && rel.Data != nil {
		details.JobProductionPlanCancellationReasonTypeID = rel.Data.ID
		reasonType = rel.Data.Type
	}

	if len(included) == 0 {
		return details
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

	if details.ChangedByID != "" && changedByType != "" {
		if user, ok := included[resourceKey(changedByType, details.ChangedByID)]; ok {
			details.ChangedBy = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	if details.JobProductionPlanCancellationReasonTypeID != "" && reasonType != "" {
		if reason, ok := included[resourceKey(reasonType, details.JobProductionPlanCancellationReasonTypeID)]; ok {
			details.JobProductionPlanCancellationReasonType = firstNonEmpty(
				strings.TrimSpace(stringAttr(reason.Attributes, "name")),
				strings.TrimSpace(stringAttr(reason.Attributes, "description")),
			)
		}
	}

	return details
}

func renderJobProductionPlanStatusChangeDetails(cmd *cobra.Command, details jobProductionPlanStatusChangeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	renderStatusChangeRelation(out, "Job Production Plan", details.JobProductionPlan, details.JobProductionPlanID)
	fmt.Fprintf(out, "Status: %s\n", details.Status)
	if details.ChangedAt != "" {
		fmt.Fprintf(out, "Changed At: %s\n", details.ChangedAt)
	}
	renderStatusChangeRelation(out, "Changed By", details.ChangedBy, details.ChangedByID)
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}
	renderStatusChangeRelation(out, "Cancellation Reason", details.JobProductionPlanCancellationReasonType, details.JobProductionPlanCancellationReasonTypeID)

	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated: %s\n", details.UpdatedAt)
	}

	return nil
}

func renderStatusChangeRelation(out io.Writer, label, name, id string) {
	if name != "" {
		if id != "" {
			fmt.Fprintf(out, "%s: %s (%s)\n", label, name, id)
		} else {
			fmt.Fprintf(out, "%s: %s\n", label, name)
		}
		return
	}
	if id != "" {
		fmt.Fprintf(out, "%s ID: %s\n", label, id)
	}
}
