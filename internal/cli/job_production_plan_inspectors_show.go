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

type jobProductionPlanInspectorsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanInspectorDetails struct {
	ID                  string `json:"id"`
	JobProductionPlanID string `json:"job_production_plan_id,omitempty"`
	UserID              string `json:"user_id,omitempty"`
}

func newJobProductionPlanInspectorsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan inspector details",
		Long: `Show the full details of a job production plan inspector.

Output Fields:
  ID                  Job production plan inspector identifier
  Job Production Plan Job production plan ID
  User                User ID

Arguments:
  <id>    The job production plan inspector ID (required). You can find IDs using the list command.`,
		Example: `  # Show a job production plan inspector
  xbe view job-production-plan-inspectors show 123

  # Get JSON output
  xbe view job-production-plan-inspectors show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanInspectorsShow,
	}
	initJobProductionPlanInspectorsShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanInspectorsCmd.AddCommand(newJobProductionPlanInspectorsShowCmd())
}

func initJobProductionPlanInspectorsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanInspectorsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanInspectorsShowOptions(cmd)
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
		return fmt.Errorf("job production plan inspector id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-inspectors/"+id, query)
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

	details := buildJobProductionPlanInspectorDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanInspectorDetails(cmd, details)
}

func parseJobProductionPlanInspectorsShowOptions(cmd *cobra.Command) (jobProductionPlanInspectorsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanInspectorsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanInspectorDetails(resp jsonAPISingleResponse) jobProductionPlanInspectorDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := jobProductionPlanInspectorDetails{
		ID:                  resource.ID,
		JobProductionPlanID: stringAttr(attrs, "job-production-plan-id"),
	}

	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
	}

	return details
}

func renderJobProductionPlanInspectorDetails(cmd *cobra.Command, details jobProductionPlanInspectorDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan: %s\n", details.JobProductionPlanID)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User: %s\n", details.UserID)
	}

	return nil
}
