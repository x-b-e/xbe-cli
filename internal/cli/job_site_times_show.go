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

type jobSiteTimesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobSiteTimeDetails struct {
	ID                  string  `json:"id"`
	JobProductionPlanID string  `json:"job_production_plan_id,omitempty"`
	UserID              string  `json:"user_id,omitempty"`
	StartAt             string  `json:"start_at,omitempty"`
	EndAt               string  `json:"end_at,omitempty"`
	Hours               float64 `json:"hours,omitempty"`
	Description         string  `json:"description,omitempty"`
	CreatedAt           string  `json:"created_at,omitempty"`
	UpdatedAt           string  `json:"updated_at,omitempty"`
}

func newJobSiteTimesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job site time details",
		Long: `Show the full details of a job site time.

Output Fields:
  ID                  Job site time identifier
  Job Production Plan Associated job production plan ID
  User                User ID
  Start At            Start timestamp
  End At              End timestamp
  Hours               Calculated duration in hours
  Description         Description
  Created At          Created timestamp
  Updated At          Updated timestamp

Arguments:
  <id>    The job site time ID (required). You can find IDs using the list command.`,
		Example: `  # Show a job site time
  xbe view job-site-times show 123

  # Get JSON output
  xbe view job-site-times show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobSiteTimesShow,
	}
	initJobSiteTimesShowFlags(cmd)
	return cmd
}

func init() {
	jobSiteTimesCmd.AddCommand(newJobSiteTimesShowCmd())
}

func initJobSiteTimesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobSiteTimesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobSiteTimesShowOptions(cmd)
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
		return fmt.Errorf("job site time id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/job-site-times/"+id, nil)
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

	details := buildJobSiteTimeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobSiteTimeDetails(cmd, details)
}

func parseJobSiteTimesShowOptions(cmd *cobra.Command) (jobSiteTimesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobSiteTimesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobSiteTimeDetails(resp jsonAPISingleResponse) jobSiteTimeDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := jobSiteTimeDetails{
		ID:          resource.ID,
		StartAt:     formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:       formatDateTime(stringAttr(attrs, "end-at")),
		Hours:       floatAttr(attrs, "hours"),
		Description: stringAttr(attrs, "description"),
		CreatedAt:   formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:   formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
	}

	return details
}

func renderJobSiteTimeDetails(cmd *cobra.Command, details jobSiteTimeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan: %s\n", details.JobProductionPlanID)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User: %s\n", details.UserID)
	}
	if details.StartAt != "" {
		fmt.Fprintf(out, "Start At: %s\n", details.StartAt)
	}
	if details.EndAt != "" {
		fmt.Fprintf(out, "End At: %s\n", details.EndAt)
	}
	if details.Hours != 0 {
		fmt.Fprintf(out, "Hours: %s\n", formatHours(details.Hours))
	}
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
