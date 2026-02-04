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

type jobScheduleShiftStartSiteChangesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobScheduleShiftStartSiteChangeDetails struct {
	ID               string `json:"id"`
	JobScheduleShift string `json:"job_schedule_shift_id,omitempty"`
	OldStartSiteType string `json:"old_start_site_type,omitempty"`
	OldStartSiteID   string `json:"old_start_site_id,omitempty"`
	NewStartSiteType string `json:"new_start_site_type,omitempty"`
	NewStartSiteID   string `json:"new_start_site_id,omitempty"`
	CreatedByID      string `json:"created_by_id,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
}

func newJobScheduleShiftStartSiteChangesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job schedule shift start site change details",
		Long: `Show the full details of a job schedule shift start site change.

Output Fields:
  ID            Start site change identifier
  SHIFT         Job schedule shift ID
  OLD START     Previous start site (type/id)
  NEW START     New start site (type/id)
  CREATED BY    User who created the change
  CREATED AT    When the change was created
  UPDATED AT    When the change was last updated

Arguments:
  <id>  The start site change ID (required).`,
		Example: `  # Show a start site change
  xbe view job-schedule-shift-start-site-changes show 123

  # JSON output
  xbe view job-schedule-shift-start-site-changes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobScheduleShiftStartSiteChangesShow,
	}
	initJobScheduleShiftStartSiteChangesShowFlags(cmd)
	return cmd
}

func init() {
	jobScheduleShiftStartSiteChangesCmd.AddCommand(newJobScheduleShiftStartSiteChangesShowCmd())
}

func initJobScheduleShiftStartSiteChangesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobScheduleShiftStartSiteChangesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseJobScheduleShiftStartSiteChangesShowOptions(cmd)
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
		return fmt.Errorf("job schedule shift start site change id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/job-schedule-shift-start-site-changes/"+id, nil)
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

	details := buildJobScheduleShiftStartSiteChangeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobScheduleShiftStartSiteChangeDetails(cmd, details)
}

func parseJobScheduleShiftStartSiteChangesShowOptions(cmd *cobra.Command) (jobScheduleShiftStartSiteChangesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobScheduleShiftStartSiteChangesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobScheduleShiftStartSiteChangeDetails(resp jsonAPISingleResponse) jobScheduleShiftStartSiteChangeDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := jobScheduleShiftStartSiteChangeDetails{
		ID:        resource.ID,
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["job-schedule-shift"]; ok && rel.Data != nil {
		details.JobScheduleShift = rel.Data.ID
	}
	if rel, ok := resource.Relationships["old-start-site"]; ok && rel.Data != nil {
		details.OldStartSiteType = rel.Data.Type
		details.OldStartSiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["new-start-site"]; ok && rel.Data != nil {
		details.NewStartSiteType = rel.Data.Type
		details.NewStartSiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderJobScheduleShiftStartSiteChangeDetails(cmd *cobra.Command, details jobScheduleShiftStartSiteChangeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobScheduleShift != "" {
		fmt.Fprintf(out, "Shift: %s\n", details.JobScheduleShift)
	}
	oldStart := formatResourceRef(details.OldStartSiteType, details.OldStartSiteID)
	if oldStart != "" {
		fmt.Fprintf(out, "Old Start Site: %s\n", oldStart)
	}
	newStart := formatResourceRef(details.NewStartSiteType, details.NewStartSiteID)
	if newStart != "" {
		fmt.Fprintf(out, "New Start Site: %s\n", newStart)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
