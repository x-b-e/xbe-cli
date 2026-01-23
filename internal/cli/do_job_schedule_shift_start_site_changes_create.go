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

type doJobScheduleShiftStartSiteChangesCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	JobScheduleShift string
	NewStartSiteType string
	NewStartSiteID   string
}

func newDoJobScheduleShiftStartSiteChangesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job schedule shift start site change",
		Long: `Create a job schedule shift start site change.

Required flags:
  --job-schedule-shift   Job schedule shift ID (required)
  --new-start-site-type  New start site type (job-sites or material-sites) (required)
  --new-start-site-id    New start site ID (required)`,
		Example: `  # Change a shift start site to a material site
  xbe do job-schedule-shift-start-site-changes create \\
    --job-schedule-shift 123 \\
    --new-start-site-type material-sites \\
    --new-start-site-id 456

  # Change a shift start site to a job site
  xbe do job-schedule-shift-start-site-changes create \\
    --job-schedule-shift 123 \\
    --new-start-site-type job-sites \\
    --new-start-site-id 789

  # JSON output
  xbe do job-schedule-shift-start-site-changes create \\
    --job-schedule-shift 123 \\
    --new-start-site-type material-sites \\
    --new-start-site-id 456 \\
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoJobScheduleShiftStartSiteChangesCreate,
	}
	initDoJobScheduleShiftStartSiteChangesCreateFlags(cmd)
	return cmd
}

func init() {
	doJobScheduleShiftStartSiteChangesCmd.AddCommand(newDoJobScheduleShiftStartSiteChangesCreateCmd())
}

func initDoJobScheduleShiftStartSiteChangesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-schedule-shift", "", "Job schedule shift ID (required)")
	cmd.Flags().String("new-start-site-type", "", "New start site type (job-sites or material-sites) (required)")
	cmd.Flags().String("new-start-site-id", "", "New start site ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobScheduleShiftStartSiteChangesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobScheduleShiftStartSiteChangesCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if opts.JobScheduleShift == "" {
		err := fmt.Errorf("--job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NewStartSiteType == "" {
		err := fmt.Errorf("--new-start-site-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NewStartSiteID == "" {
		err := fmt.Errorf("--new-start-site-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"job-schedule-shift": map[string]any{
			"data": map[string]any{
				"type": "job-schedule-shifts",
				"id":   opts.JobScheduleShift,
			},
		},
		"new-start-site": map[string]any{
			"data": map[string]any{
				"type": opts.NewStartSiteType,
				"id":   opts.NewStartSiteID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-schedule-shift-start-site-changes",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-schedule-shift-start-site-changes", jsonBody)
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

	row := buildJobScheduleShiftStartSiteChangeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.JobScheduleShift != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created job schedule shift start site change %s for shift %s\n", row.ID, row.JobScheduleShift)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job schedule shift start site change %s\n", row.ID)
	return nil
}

func parseDoJobScheduleShiftStartSiteChangesCreateOptions(cmd *cobra.Command) (doJobScheduleShiftStartSiteChangesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobScheduleShift, _ := cmd.Flags().GetString("job-schedule-shift")
	newStartSiteType, _ := cmd.Flags().GetString("new-start-site-type")
	newStartSiteID, _ := cmd.Flags().GetString("new-start-site-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobScheduleShiftStartSiteChangesCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		JobScheduleShift: jobScheduleShift,
		NewStartSiteType: newStartSiteType,
		NewStartSiteID:   newStartSiteID,
	}, nil
}

func buildJobScheduleShiftStartSiteChangeRowFromSingle(resp jsonAPISingleResponse) jobScheduleShiftStartSiteChangeRow {
	return buildJobScheduleShiftStartSiteChangeRow(resp.Data)
}
